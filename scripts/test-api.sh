#!/bin/bash

# API Testing Script for Dear Future
# Tests all API endpoints with example requests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API Base URL
API_URL="${API_URL:-http://localhost:8080}"

# Test credentials
TEST_EMAIL="test-$(date +%s)@example.com"
TEST_PASSWORD="TestPassword123"
TEST_NAME="Test User"

echo -e "${BLUE}🧪 Testing Dear Future API${NC}"
echo -e "${BLUE}API URL: $API_URL${NC}\n"

# Function to make HTTP request and show result
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4
    local description=$5

    echo -e "${YELLOW}Testing: $description${NC}"
    echo -e "${BLUE}$method $endpoint${NC}"

    if [ -n "$token" ]; then
        if [ -n "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$API_URL$endpoint" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $token" \
                -d "$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$API_URL$endpoint" \
                -H "Authorization: Bearer $token")
        fi
    else
        if [ -n "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$API_URL$endpoint" \
                -H "Content-Type: application/json" \
                -d "$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$API_URL$endpoint")
        fi
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ Success ($http_code)${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}❌ Failed ($http_code)${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    fi
    echo ""

    # Return the body for later use
    echo "$body"
}

# 1. Test Health Check
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}1. Health & Info Endpoints${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

test_endpoint "GET" "/health" "" "" "Health Check" > /dev/null
test_endpoint "GET" "/api/v1/" "" "" "API Information" > /dev/null

# 2. Test User Registration
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}2. User Registration${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

REGISTER_DATA="{
  \"email\": \"$TEST_EMAIL\",
  \"name\": \"$TEST_NAME\",
  \"password\": \"$TEST_PASSWORD\",
  \"timezone\": \"America/New_York\"
}"

REGISTER_RESPONSE=$(test_endpoint "POST" "/api/v1/auth/register" "$REGISTER_DATA" "" "Register New User")
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token' 2>/dev/null)
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.refresh_token' 2>/dev/null)

if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
    echo -e "${GREEN}✅ Got access token: ${ACCESS_TOKEN:0:20}...${NC}\n"
else
    echo -e "${RED}❌ Failed to get access token${NC}\n"
    exit 1
fi

# 3. Test User Login
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}3. User Login${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

LOGIN_DATA="{
  \"email\": \"$TEST_EMAIL\",
  \"password\": \"$TEST_PASSWORD\"
}"

LOGIN_RESPONSE=$(test_endpoint "POST" "/api/v1/auth/login" "$LOGIN_DATA" "" "Login with Credentials")
NEW_ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token' 2>/dev/null)

if [ "$NEW_ACCESS_TOKEN" != "null" ] && [ -n "$NEW_ACCESS_TOKEN" ]; then
    ACCESS_TOKEN="$NEW_ACCESS_TOKEN"
    echo -e "${GREEN}✅ Login successful${NC}\n"
fi

# 4. Test User Profile
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}4. User Profile${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

test_endpoint "GET" "/api/v1/user/profile" "" "$ACCESS_TOKEN" "Get User Profile" > /dev/null

# 5. Test Update Profile
UPDATE_DATA="{
  \"name\": \"Updated Test User\",
  \"timezone\": \"Europe/London\"
}"

test_endpoint "PUT" "/api/v1/user/update" "$UPDATE_DATA" "$ACCESS_TOKEN" "Update User Profile" > /dev/null

# 6. Test Message Creation
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}5. Message Operations${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

FUTURE_DATE=$(date -u -v+1y +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 year" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2026-01-01T00:00:00Z")

MESSAGE_DATA="{
  \"title\": \"Test Message to Future Self\",
  \"content\": \"Hello future me! This is a test message created via API.\",
  \"delivery_date\": \"$FUTURE_DATE\",
  \"timezone\": \"UTC\",
  \"delivery_method\": \"email\"
}"

CREATE_RESPONSE=$(test_endpoint "POST" "/api/v1/messages" "$MESSAGE_DATA" "$ACCESS_TOKEN" "Create Message")
MESSAGE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id' 2>/dev/null)

if [ "$MESSAGE_ID" != "null" ] && [ -n "$MESSAGE_ID" ]; then
    echo -e "${GREEN}✅ Message created with ID: $MESSAGE_ID${NC}\n"
fi

# 7. Test Get All Messages
test_endpoint "GET" "/api/v1/messages?limit=10" "" "$ACCESS_TOKEN" "Get All Messages" > /dev/null

# 8. Test Get Single Message
if [ "$MESSAGE_ID" != "null" ] && [ -n "$MESSAGE_ID" ]; then
    test_endpoint "GET" "/api/v1/messages?id=$MESSAGE_ID" "" "$ACCESS_TOKEN" "Get Single Message" > /dev/null
fi

# 9. Test Update Message
if [ "$MESSAGE_ID" != "null" ] && [ -n "$MESSAGE_ID" ]; then
    UPDATE_MESSAGE_DATA="{
      \"title\": \"Updated Test Message\",
      \"content\": \"This message has been updated!\"
    }"
    test_endpoint "PUT" "/api/v1/messages?id=$MESSAGE_ID" "$UPDATE_MESSAGE_DATA" "$ACCESS_TOKEN" "Update Message" > /dev/null
fi

# 10. Test Refresh Token
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}6. Token Refresh${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

if [ "$REFRESH_TOKEN" != "null" ] && [ -n "$REFRESH_TOKEN" ]; then
    REFRESH_DATA="{
      \"refresh_token\": \"$REFRESH_TOKEN\"
    }"
    test_endpoint "POST" "/api/v1/auth/refresh" "$REFRESH_DATA" "" "Refresh Access Token" > /dev/null
fi

# 11. Test Delete Message
if [ "$MESSAGE_ID" != "null" ] && [ -n "$MESSAGE_ID" ]; then
    echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}7. Delete Message${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

    test_endpoint "DELETE" "/api/v1/messages?id=$MESSAGE_ID" "" "$ACCESS_TOKEN" "Delete Message" > /dev/null
fi

# 12. Test Authentication Failure
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}8. Authentication Tests${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

test_endpoint "GET" "/api/v1/user/profile" "" "invalid_token" "Test Invalid Token (should fail)" > /dev/null
test_endpoint "GET" "/api/v1/user/profile" "" "" "Test No Token (should fail)" > /dev/null

# Summary
echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${GREEN}✅ API Testing Complete!${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}\n"

echo -e "${GREEN}Summary:${NC}"
echo -e "  • Registered user: $TEST_EMAIL"
echo -e "  • Access token obtained"
echo -e "  • Profile operations tested"
echo -e "  • Message CRUD operations tested"
echo -e "  • Authentication tested"
echo ""
echo -e "${YELLOW}Note: Check server logs for detailed information${NC}"
echo -e "${YELLOW}Server logs: /tmp/server-test.log${NC}\n"
