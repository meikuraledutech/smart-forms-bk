#!/bin/bash

BASE_URL="http://localhost:3030"

# Register and login
echo "=== Register User ==="
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}' > /dev/null

echo "=== Login ==="
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}' | jq -r '.access_token')

echo "Token: ${TOKEN:0:30}..."

# Create form
echo -e "\n=== Create Form ==="
FORM_ID=$(curl -s -X POST "$BASE_URL/forms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Test Form","description":"Test"}' | jq -r '.id')

echo "Form ID: $FORM_ID"

# Create flow
echo -e "\n=== Create Flow ==="
curl -s -X PATCH "$BASE_URL/forms/$FORM_ID/flow" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "blocks": [
      {
        "id": "q1",
        "type": "question",
        "question": "What is your name?",
        "children": []
      }
    ]
  }' > /dev/null

# Publish form
echo -e "\n=== Publish Form ==="
SLUG_DATA=$(curl -s -X PATCH "$BASE_URL/forms/$FORM_ID/publish" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN")

SLUG=$(echo "$SLUG_DATA" | jq -r '.links.auto_slug')
echo "Slug: $SLUG"

# Get flow connection ID
echo -e "\n=== Get Flow ==="
FLOW_CONN_ID=$(curl -s -X GET "$BASE_URL/forms/$FORM_ID/flow" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.blocks[0].id')

echo "Flow Connection ID: $FLOW_CONN_ID"

# Submit response
echo -e "\n=== Submit Response ==="
RESPONSE_ID=$(curl -s -X POST "$BASE_URL/f/$SLUG/responses" \
  -H "Content-Type: application/json" \
  -d "{
    \"responses\": [
      {
        \"flow_connection_id\": \"$FLOW_CONN_ID\",
        \"answer_text\": \"John Doe\",
        \"time_spent\": 5
      }
    ],
    \"metadata\": {
      \"total_time_spent\": 10,
      \"flow_path\": [\"$FLOW_CONN_ID\"]
    }
  }" | jq -r '.response_id')

echo "Response ID: $RESPONSE_ID"

# Get response details
echo -e "\n=== Get Individual Response Details ==="
curl -s -X GET "$BASE_URL/responses/$RESPONSE_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo -e "\n=== Test Complete ==="
