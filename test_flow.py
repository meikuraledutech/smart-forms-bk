#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:3030"

def main():
    print("=== Test Flow API ===\n")

    # Step 1: Login
    print("1. Login as user...")
    login_response = requests.post(
        f"{BASE_URL}/auth/login",
        json={"username": "usera", "password": "test123"}
    )

    if login_response.status_code != 200:
        print(f"Login failed: {login_response.text}")
        return

    token = login_response.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}
    print(f"✅ Login successful\n")

    # Step 2: Create form
    print("2. Create form...")
    form_response = requests.post(
        f"{BASE_URL}/forms",
        headers=headers,
        json={"title": "Flow Test Form", "description": "Testing flow tree reconstruction"}
    )

    if form_response.status_code not in [200, 201]:
        print(f"Create form failed: {form_response.text}")
        return

    form_id = form_response.json()["id"]
    print(f"✅ Form created: {form_id}\n")

    # Step 3: Create flow
    print("3. Create flow...")
    flow_data = {
        "blocks": [
            {
                "id": "1766738655070",
                "type": "question",
                "question": "Department",
                "children": [
                    {
                        "id": "1766739095856",
                        "type": "option",
                        "question": "Mechanical",
                        "children": [
                            {
                                "id": "1766739100001",
                                "type": "question",
                                "question": "What is your employee ID?",
                                "children": []
                            }
                        ]
                    },
                    {
                        "id": "1766739095857",
                        "type": "option",
                        "question": "Computer Science",
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
        print(f"Create flow failed: {create_flow_response.text}")
        return

    print("✅ Flow created successfully")
    print("ID Mapping:")
    print(json.dumps(create_flow_response.json()["mapping"], indent=2))
    print()

    # Step 4: Get flow tree
    print("4. Get flow tree...")
    get_flow_response = requests.get(
        f"{BASE_URL}/forms/{form_id}/flow",
        headers=headers
    )

    if get_flow_response.status_code != 200:
        print(f"Get flow failed: {get_flow_response.text}")
        return

    print("✅ Flow retrieved successfully")
    print("Tree structure:")
    print(json.dumps(get_flow_response.json(), indent=2))
    print()

    # Verify structure
    print("5. Verify tree structure...")
    retrieved_blocks = get_flow_response.json()["blocks"]

    if retrieved_blocks and len(retrieved_blocks) > 0:
        root_block = retrieved_blocks[0]
        print(f"✅ Root question: {root_block['question']}")
        print(f"✅ Root type: {root_block['type']}")
        print(f"✅ Children count: {len(root_block['children'])}")

        if root_block['children'] and len(root_block['children']) > 0:
            first_child = root_block['children'][0]
            print(f"✅ First child question: {first_child['question']}")
            print(f"✅ First child type: {first_child['type']}")

            if first_child['children'] and len(first_child['children']) > 0:
                nested_child = first_child['children'][0]
                print(f"✅ Nested child question: {nested_child['question']}")
                print(f"✅ Nested child type: {nested_child['type']}")
                print("\n✅ Tree reconstruction working perfectly!")
            else:
                print("⚠️ Nested children missing")
        else:
            print("⚠️ Children missing")
    else:
        print("❌ No blocks retrieved")

if __name__ == "__main__":
    main()
