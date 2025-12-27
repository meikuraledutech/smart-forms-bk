#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:3030"

def main():
    print("=== Test Links API ===\n")

    # Step 1: Register and login
    print("1. Register and login...")
    requests.post(f"{BASE_URL}/auth/register", json={"username": "linkuser", "password": "test123"})

    login_response = requests.post(f"{BASE_URL}/auth/login", json={"username": "linkuser", "password": "test123"})
    token = login_response.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}
    print(f"✅ Logged in\n")

    # Step 2: Create form
    print("2. Create form...")
    form_response = requests.post(
        f"{BASE_URL}/forms",
        headers=headers,
        json={"title": "Employee Survey", "description": "Annual employee feedback form"}
    )
    form_id = form_response.json()["id"]
    print(f"✅ Form created: {form_id}\n")

    # Step 3: Create flow
    print("3. Create flow...")
    flow_data = {
        "blocks": [
            {
                "id": "q1",
                "type": "question",
                "question": "How satisfied are you with your role?",
                "children": [
                    {"id": "o1", "type": "option", "question": "Very Satisfied", "children": []},
                    {"id": "o2", "type": "option", "question": "Satisfied", "children": []},
                    {"id": "o3", "type": "option", "question": "Neutral", "children": []}
                ]
            }
        ]
    }
    requests.patch(f"{BASE_URL}/forms/{form_id}/flow", headers=headers, json=flow_data)
    print("✅ Flow created\n")

    # Step 4: Publish form without custom slug
    print("4. Publish form (auto-slug only)...")
    publish_response = requests.patch(f"{BASE_URL}/forms/{form_id}/publish", headers=headers, json={})
    publish_data = publish_response.json()
    print(f"✅ Published:")
    print(json.dumps(publish_data, indent=2))
    auto_slug = publish_data["links"]["auto_slug"]
    print()

    # Step 5: Get public form via auto slug
    print(f"5. Get public form via auto slug: /f/{auto_slug}")
    public_response = requests.get(f"{BASE_URL}/f/{auto_slug}")
    print(f"Status: {public_response.status_code}")
    if public_response.status_code == 200:
        public_form = public_response.json()
        print(f"✅ Title: {public_form['title']}")
        print(f"✅ Accepting responses: {public_form['accepting_responses']}")
        print(f"✅ Flow blocks count: {len(public_form['flow']['blocks'])}")
        print()
    else:
        print(f"❌ Failed: {public_response.text}\n")

    # Step 6: Create another form with custom slug
    print("6. Create second form with custom slug...")
    form2_response = requests.post(
        f"{BASE_URL}/forms",
        headers=headers,
        json={"title": "Feedback Form", "description": "General feedback"}
    )
    form2_id = form2_response.json()["id"]

    # Create flow for form 2
    requests.patch(f"{BASE_URL}/forms/{form2_id}/flow", headers=headers, json=flow_data)

    # Publish with custom slug
    publish2_response = requests.patch(
        f"{BASE_URL}/forms/{form2_id}/publish",
        headers=headers,
        json={"custom_slug": "employee-feedback"}
    )
    publish2_data = publish2_response.json()
    print(f"✅ Published with custom slug:")
    print(json.dumps(publish2_data, indent=2))
    custom_slug = publish2_data["links"]["custom_slug"]
    print()

    # Step 7: Get public form via custom slug
    print(f"7. Get public form via custom slug: /f/{custom_slug}")
    custom_response = requests.get(f"{BASE_URL}/f/{custom_slug}")
    if custom_response.status_code == 200:
        print(f"✅ Accessible via custom slug")
        print(f"✅ Title: {custom_response.json()['title']}\n")
    else:
        print(f"❌ Failed: {custom_response.text}\n")

    # Step 8: Toggle accepting responses off
    print("8. Toggle accepting responses OFF...")
    toggle_response = requests.patch(
        f"{BASE_URL}/forms/{form_id}/accepting-responses",
        headers=headers,
        json={"accepting": False}
    )
    print(f"✅ Response: {toggle_response.json()}\n")

    # Step 9: Verify accepting_responses is false
    print("9. Verify accepting_responses is now false...")
    verify_response = requests.get(f"{BASE_URL}/f/{auto_slug}")
    if verify_response.status_code == 200:
        is_accepting = verify_response.json()["accepting_responses"]
        if not is_accepting:
            print(f"✅ Confirmed: accepting_responses = {is_accepting}\n")
        else:
            print(f"❌ Still accepting responses\n")

    # Step 10: Test duplicate custom slug (should fail)
    print("10. Test duplicate custom slug (should fail)...")
    form3_response = requests.post(
        f"{BASE_URL}/forms",
        headers=headers,
        json={"title": "Another Form", "description": "Test"}
    )
    form3_id = form3_response.json()["id"]
    requests.patch(f"{BASE_URL}/forms/{form3_id}/flow", headers=headers, json=flow_data)

    duplicate_response = requests.patch(
        f"{BASE_URL}/forms/{form3_id}/publish",
        headers=headers,
        json={"custom_slug": "employee-feedback"}
    )
    if duplicate_response.status_code == 409:
        print(f"✅ Correctly rejected duplicate slug (409 Conflict)")
        print(f"   Message: {duplicate_response.json()}\n")
    else:
        print(f"❌ Should have rejected duplicate slug\n")

    # Step 11: Test invalid custom slug format
    print("11. Test invalid custom slug format (should fail)...")
    invalid_response = requests.patch(
        f"{BASE_URL}/forms/{form3_id}/publish",
        headers=headers,
        json={"custom_slug": "Invalid Slug!"}
    )
    if invalid_response.status_code == 400:
        print(f"✅ Correctly rejected invalid slug (400 Bad Request)")
        print(f"   Message: {invalid_response.json()}\n")
    else:
        print(f"❌ Should have rejected invalid slug\n")

    print("=== All tests completed! ===")

if __name__ == "__main__":
    main()
