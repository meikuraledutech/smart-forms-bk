#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:3030"

def main():
    print("=== Test Responses API ===\n")

    # Step 1: Register and login
    print("1. Register and login...")
    requests.post(f"{BASE_URL}/auth/register", json={"username": "responseuser", "password": "test123"})

    login_response = requests.post(f"{BASE_URL}/auth/login", json={"username": "responseuser", "password": "test123"})

    if login_response.status_code != 200:
        print(f"❌ Login failed: {login_response.text}")
        return

    token = login_response.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}
    print("✅ Logged in\n")

    # Step 2: Create form
    print("2. Create form...")
    form_response = requests.post(
        f"{BASE_URL}/forms",
        headers=headers,
        json={"title": "Customer Feedback", "description": "Tell us your experience"}
    )

    if form_response.status_code not in [200, 201]:
        print(f"❌ Create form failed: {form_response.text}")
        return

    form_id = form_response.json()["id"]
    print(f"✅ Form created: {form_id}\n")

    # Step 3: Create flow
    print("3. Create flow...")
    flow_data = {
        "blocks": [
            {
                "id": "q1",
                "type": "question",
                "question": "How satisfied are you with our service?",
                "children": [
                    {
                        "id": "opt1",
                        "type": "option",
                        "question": "Very Satisfied",
                        "children": [
                            {
                                "id": "followup1",
                                "type": "input",
                                "question": "What did you like most?",
                                "children": []
                            }
                        ]
                    },
                    {
                        "id": "opt2",
                        "type": "option",
                        "question": "Satisfied",
                        "children": []
                    },
                    {
                        "id": "opt3",
                        "type": "option",
                        "question": "Neutral",
                        "children": []
                    }
                ]
            }
        ]
    }

    create_flow_response = requests.patch(
        f"{BASE_URL}/forms/{form_id}/flow",
        headers=headers,
        json=flow_data
    )

    if create_flow_response.status_code != 200:
        print(f"❌ Create flow failed: {create_flow_response.text}")
        return

    flow_mapping = create_flow_response.json()["mapping"]
    print("✅ Flow created")
    print(f"   ID Mapping: {json.dumps(flow_mapping, indent=2)}\n")

    # Step 4: Publish form
    print("4. Publish form...")
    import time
    slug_suffix = str(int(time.time()))

    publish_response = requests.patch(
        f"{BASE_URL}/forms/{form_id}/publish",
        headers=headers,
        json={"custom_slug": f"feedback-{slug_suffix}"}
    )

    if publish_response.status_code != 200:
        print(f"❌ Publish failed: {publish_response.text}")
        return

    slug = publish_response.json()["links"]["custom_slug"]
    print(f"✅ Published with slug: {slug}\n")

    # Step 5: Get public form
    print("5. Get public form to get flow_connection_ids...")
    public_form = requests.get(f"{BASE_URL}/f/{slug}")

    if public_form.status_code != 200:
        print(f"❌ Get public form failed: {public_form.status_code}")
        return

    form_data = public_form.json()
    blocks = form_data["flow"]["blocks"]

    # Extract flow_connection_ids from the flow
    question_id = blocks[0]["id"]
    option_id = blocks[0]["children"][0]["id"]  # "Very Satisfied"
    followup_id = blocks[0]["children"][0]["children"][0]["id"]  # "What did you like most?"

    print(f"✅ Got flow structure")
    print(f"   Question ID: {question_id}")
    print(f"   Option ID: {option_id}")
    print(f"   Follow-up ID: {followup_id}\n")

    # Step 6: Submit response (PUBLIC endpoint, no auth)
    print("6. Submit response (public endpoint)...")
    submit_data = {
        "responses": [
            {
                "flow_connection_id": question_id,
                "answer_text": "How satisfied are you with our service?",
                "time_spent": 5
            },
            {
                "flow_connection_id": option_id,
                "answer_text": "Very Satisfied",
                "time_spent": 3
            },
            {
                "flow_connection_id": followup_id,
                "answer_text": "Great customer support and fast delivery!",
                "answer_value": {
                    "sentiment": "positive",
                    "length": 45
                },
                "time_spent": 15
            }
        ],
        "metadata": {
            "total_time_spent": 23,
            "flow_path": [question_id, option_id, followup_id]
        }
    }

    submit_response = requests.post(
        f"{BASE_URL}/f/{slug}/responses",
        json=submit_data
    )

    print(f"Status: {submit_response.status_code}")
    print(f"Response: {submit_response.text}")

    if submit_response.status_code == 201:
        response_id = submit_response.json()["response_id"]
        print(f"✅ Response submitted successfully")
        print(f"   Response ID: {response_id}\n")
    else:
        print(f"❌ Submit failed\n")
        return

    # Step 7: Submit another response
    print("7. Submit second response...")
    submit_data2 = {
        "responses": [
            {
                "flow_connection_id": question_id,
                "answer_text": "How satisfied are you with our service?",
                "time_spent": 4
            },
            {
                "flow_connection_id": blocks[0]["children"][1]["id"],  # "Satisfied"
                "answer_text": "Satisfied",
                "time_spent": 2
            }
        ],
        "metadata": {
            "total_time_spent": 6,
            "flow_path": [question_id, blocks[0]["children"][1]["id"]]
        }
    }

    submit_response2 = requests.post(
        f"{BASE_URL}/f/{slug}/responses",
        json=submit_data2
    )

    if submit_response2.status_code == 201:
        print(f"✅ Second response submitted\n")
    else:
        print(f"❌ Second submit failed: {submit_response2.text}\n")

    # Step 8: Get responses (protected endpoint)
    print("8. Get all responses (owner only)...")
    get_responses = requests.get(
        f"{BASE_URL}/forms/{form_id}/responses",
        headers=headers
    )

    print(f"Status: {get_responses.status_code}")

    if get_responses.status_code == 200:
        data = get_responses.json()
        print(f"✅ Retrieved responses")
        print(f"   Total: {data['total']}")
        print(f"   Items: {len(data['items'])}")
        print("\n   Response details:")
        for i, item in enumerate(data['items'], 1):
            print(f"   {i}. ID: {item['id']}")
            print(f"      Submitted: {item['submitted_at']}")
            print(f"      Time spent: {item['total_time_spent']}s")
            print(f"      Flow path length: {len(item['flow_path'])}")
        print()
    else:
        print(f"❌ Get responses failed: {get_responses.text}\n")

    # Step 9: Test validation - form not accepting responses
    print("9. Test validation: Stop accepting responses...")
    toggle_response = requests.patch(
        f"{BASE_URL}/forms/{form_id}/accepting-responses",
        headers=headers,
        json={"accepting": False}
    )

    if toggle_response.status_code == 200:
        print("✅ Stopped accepting responses\n")

    print("10. Try to submit when not accepting...")
    submit_rejected = requests.post(
        f"{BASE_URL}/f/{slug}/responses",
        json=submit_data
    )

    print(f"Status: {submit_rejected.status_code}")
    if submit_rejected.status_code == 403:
        print(f"✅ Correctly rejected: {submit_rejected.text}\n")
    else:
        print(f"❌ Should have rejected with 403\n")

    # Step 11: Test with invalid flow_connection_id
    print("11. Test validation: Invalid flow_connection_id...")
    toggle_on = requests.patch(
        f"{BASE_URL}/forms/{form_id}/accepting-responses",
        headers=headers,
        json={"accepting": True}
    )

    invalid_data = {
        "responses": [
            {
                "flow_connection_id": "invalid-uuid-here",
                "answer_text": "Test",
                "time_spent": 5
            }
        ],
        "metadata": {
            "total_time_spent": 5,
            "flow_path": ["invalid-uuid-here"]
        }
    }

    invalid_response = requests.post(
        f"{BASE_URL}/f/{slug}/responses",
        json=invalid_data
    )

    print(f"Status: {invalid_response.status_code}")
    if invalid_response.status_code == 400:
        print(f"✅ Correctly rejected invalid flow_connection_id\n")
    else:
        print(f"❌ Should have rejected with 400\n")

    print("=== All tests completed! ===")

if __name__ == "__main__":
    main()
