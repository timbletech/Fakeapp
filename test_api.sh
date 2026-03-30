#!/bin/bash

# Configuration
API_BASE="http://localhost:8097"
TIMBLE_API_KEY="timble-test-key-2026"
CLIENT_ID="client_123"
USER_REF="user_457"
DEVICE_ID="dev_789"
PUBLIC_KEY="LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFTnZMUXgyaXRReWdoV1RzWFgwajlLRzdvUE4zLwpybUZBNXBLY2h2VFQrU3BTWWhyQm5iMlo1clRwZ0lLVW9IWFUzTEM0aGhOK2NJaDIrKzBHaHdnSHlnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="

echo "--- 1. Testing Device Registration ---"
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/v1/device/register" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"user_ref\": \"$USER_REF\",
       \"device_info\": {
         \"device_id\": \"$DEVICE_ID\",
         \"platform\": \"ios\",
         \"device_model\": \"iPhone 15\",
         \"os_version\": \"17.0\"
       },
       \"public_key\": \"$PUBLIC_KEY\"
     }")

echo "$REGISTER_RESPONSE"
BINDING_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.device_binding_id')
if [ "$BINDING_ID" == "null" ] || [ -z "$BINDING_ID" ]; then
  echo "Error: Failed to register device"
  exit 1
fi
echo "Captured Binding ID: $BINDING_ID"

echo -e "\n--- 2. Testing Device Check ---"
curl -s -G "$API_BASE/v1/device/check" \
     --data-urlencode "client_id=$CLIENT_ID" \
     --data-urlencode "user_ref=$USER_REF" \
     --data-urlencode "device_id=$DEVICE_ID" | jq .

echo -e "\n--- 3. Testing Device Update ---"
curl -s -X PUT "$API_BASE/v1/device/update" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"user_ref\": \"$USER_REF\",
       \"device_info\": {
         \"device_id\": \"$DEVICE_ID\",
         \"platform\": \"ios\",
         \"os_version\": \"17.1\"
       },
       \"public_key\": \"$PUBLIC_KEY\"
     }" | jq .

echo -e "\n--- 4. Testing Auth Start ---"
AUTH_START_RESP=$(curl -s -X POST "$API_BASE/v1/auth/start" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"user_ref\": \"$USER_REF\",
       \"action\": \"login\",
       \"mode\": \"device\",
       \"device_binding_id\": \"$BINDING_ID\",
       \"device_info\": {
         \"device_id\": \"$DEVICE_ID\",
         \"platform\": \"ios\"
       }
     }")

echo "$AUTH_START_RESP" | jq .
SESSION_ID=$(echo "$AUTH_START_RESP" | jq -r '.auth_session_id')
CHALLENGE_ID=$(echo "$AUTH_START_RESP" | jq -r '.device.challenge_id')
CHALLENGE=$(echo "$AUTH_START_RESP" | jq -r '.device.challenge')

echo -e "\n--- 5. Generating Device Signature ---"
# Using the devsigner tool to sign the challenge
SIGNATURE=$(go run cmd/devsigner/main.go sign --challenge "$CHALLENGE")
echo "Generated Signature: $SIGNATURE"

echo -e "\n--- 6. Testing Auth Complete ---"
AUTH_COMP_RESP=$(curl -s -X POST "$API_BASE/v1/auth/complete" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"mode\": \"device\",
       \"auth_session_id\": \"$SESSION_ID\",
       \"challenge_id\": \"$CHALLENGE_ID\",
       \"device_signature\": \"$SIGNATURE\"
     }")

echo "$AUTH_COMP_RESP" | jq .
CONTEXT_TOKEN=$(echo "$AUTH_COMP_RESP" | jq -r '.auth_context_token')

echo -e "\n--- 7. Testing Auth Verify ---"
curl -s -X POST "$API_BASE/v1/auth/verify" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"auth_context_token\": \"$CONTEXT_TOKEN\"
     }" | jq .

echo -e "\n--- 8. Testing Unified Auth Start (SIM mode) ---"
SIM_UNIFIED_START_RESP=$(curl -s -X POST "$API_BASE/v1/auth/start" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"user_ref\": \"$USER_REF\",
       \"action\": \"check\",
       \"mode\": \"sim\",
       \"msisdn\": \"917905968734\"
     }")
echo "$SIM_UNIFIED_START_RESP" | jq .
SIM_UNIFIED_SESSION_ID=$(echo "$SIM_UNIFIED_START_RESP" | jq -r '.auth_session_id')

echo -e "\n--- 9. Testing Unified Auth Complete (SIM mode) ---"
if [ "$SIM_UNIFIED_SESSION_ID" != "null" ] && [ -n "$SIM_UNIFIED_SESSION_ID" ]; then
    curl -s -X POST "$API_BASE/v1/auth/complete" \
         -H "Content-Type: application/json" \
         -d "{
           \"client_id\": \"$CLIENT_ID\",
           \"mode\": \"sim\",
           \"auth_session_id\": \"$SIM_UNIFIED_SESSION_ID\"
         }" | jq .
else
    echo "Skipping unified SIM complete as auth start failed"
fi

echo -e "\n--- 10. Testing Hybrid Start ---"
HYBRID_START_RESP=$(curl -s -X POST "$API_BASE/v1/auth/start" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"user_ref\": \"$USER_REF\",
       \"action\": \"login\",
       \"mode\": \"hybrid\",
       \"msisdn\": \"917905968734\",
       \"device_binding_id\": \"$BINDING_ID\",
       \"device_info\": {
         \"device_id\": \"$DEVICE_ID\",
         \"platform\": \"ios\"
       }
     }")
echo "$HYBRID_START_RESP" | jq .
HYBRID_SESSION_ID=$(echo "$HYBRID_START_RESP" | jq -r '.auth_session_id')
HYBRID_CHALLENGE=$(echo "$HYBRID_START_RESP" | jq -r '.device.challenge')

echo -e "\n--- 11. Testing Hybrid Complete ---"
if [ "$HYBRID_SESSION_ID" != "null" ] && [ -n "$HYBRID_SESSION_ID" ] && [ "$HYBRID_CHALLENGE" != "null" ] && [ -n "$HYBRID_CHALLENGE" ]; then
    HYBRID_SIGNATURE=$(go run cmd/devsigner/main.go sign --challenge "$HYBRID_CHALLENGE")
    curl -s -X POST "$API_BASE/v1/auth/complete" \
         -H "Content-Type: application/json" \
         -d "{
           \"client_id\": \"$CLIENT_ID\",
           \"mode\": \"hybrid\",
           \"auth_session_id\": \"$HYBRID_SESSION_ID\",
           \"device_signature\": \"$HYBRID_SIGNATURE\"
         }" | jq .
else
    echo "Skipping hybrid complete as hybrid start failed"
fi

echo -e "\n--- 12. Testing Direct SIM Start (Legacy Route) ---"
SIM_START_RESP=$(curl -s -X POST "$API_BASE/v1/sim/start" \
     -H "Content-Type: application/json" \
     -H "X-API-Key: $TIMBLE_API_KEY" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"user_ref\": \"$USER_REF\",
       \"msisdn\": \"917905968734\",
       \"action\": \"check\"
     }")
echo "$SIM_START_RESP" | jq .
SIM_SESSION_ID=$(echo "$SIM_START_RESP" | jq -r '.auth_session_id')

echo -e "\n--- 13. Testing Direct SIM Complete (Legacy Route) ---"
if [ "$SIM_SESSION_ID" != "null" ] && [ -n "$SIM_SESSION_ID" ]; then
    curl -s -X POST "$API_BASE/v1/sim/complete" \
         -H "Content-Type: application/json" \
         -H "X-API-Key: $TIMBLE_API_KEY" \
         -d "{
           \"auth_session_id\": \"$SIM_SESSION_ID\"
         }" | jq .
else
    echo "Skipping SIM complete as SIM start failed (Upstream might be down)"
fi

echo -e "\n--- 14. Testing Device Revoke ---"
curl -s -X POST "$API_BASE/v1/device/revoke" \
     -H "Content-Type: application/json" \
     -d "{
       \"client_id\": \"$CLIENT_ID\",
       \"user_ref\": \"$USER_REF\",
       \"device_binding_id\": \"$BINDING_ID\"
     }" | jq .

echo -e "\nFinal Check: Device should be inactive now"
curl -s -G "$API_BASE/v1/device/check" \
     --data-urlencode "client_id=$CLIENT_ID" \
     --data-urlencode "user_ref=$USER_REF" \
     --data-urlencode "device_id=$DEVICE_ID" | jq .

echo -e "\nTest Suite Completed"
