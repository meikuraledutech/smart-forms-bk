#!/bin/bash

BASE_URL="http://localhost:3030"

# User A credentials
echo "=== Login User A ==="
TOKEN_A=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"userA","password":"password123"}' | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

echo "Token A: ${TOKEN_A:0:20}..."

# Get first form
echo -e "\n=== Get User A Forms ==="
FORM_ID=$(curl -s -X GET "$BASE_URL/forms" \
  -H "Authorization: Bearer $TOKEN_A" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

echo "Form ID: $FORM_ID"

# Get responses for form
echo -e "\n=== Get Responses for Form ==="
RESPONSE_DATA=$(curl -s -X GET "$BASE_URL/forms/$FORM_ID/responses" \
  -H "Authorization: Bearer $TOKEN_A")

echo "$RESPONSE_DATA" | head -20

# Extract first response ID
RESPONSE_ID=$(echo "$RESPONSE_DATA" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

echo -e "\nResponse ID: $RESPONSE_ID"

# Get individual response details
echo -e "\n=== Get Individual Response Details ==="
curl -s -X GET "$BASE_URL/responses/$RESPONSE_ID" \
  -H "Authorization: Bearer $TOKEN_A" | head -50

echo -e "\n\n=== Test Complete ==="
