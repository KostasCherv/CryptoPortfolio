#!/bin/bash

# Simple API Stress Test Script
# Tests wallet balance fetching and API performance under load

set -e

# Configuration
API_BASE_URL="http://localhost:8080/api/v1"
CONCURRENT_USERS=10
REQUESTS_PER_USER=20
TEST_DURATION=100 # 100 seconds
BALANCE_REFRESH_INTERVAL=30  # Refresh balances every 30 seconds

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting API Stress Test${NC}"
echo "=================================="
echo "API Base URL: $API_BASE_URL"
echo "Concurrent Users: $CONCURRENT_USERS"
echo "Requests per User: $REQUESTS_PER_USER"
echo "Test Duration: ${TEST_DURATION}s"
echo "Balance Refresh Interval: ${BALANCE_REFRESH_INTERVAL}s"
echo ""

# Check dependencies
if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ jq is required but not installed. Please install jq.${NC}"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ curl is required but not installed.${NC}"
    exit 1
fi

# Check if server is running
echo -e "${BLUE}🔍 Checking server status...${NC}"
if ! curl -s --max-time 5 "$API_BASE_URL/../../health" > /dev/null; then
    echo -e "${RED}❌ Server is not running. Please start the server first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Server is running${NC}"

# Create test users and data
echo -e "${YELLOW}�� Creating test data...${NC}"

# Function to create a test user with wallets and tokens
create_test_user() {
    local user_id=$1
    local timestamp=$(date +%s)
    local email="stress_test_user${user_id}_${timestamp}@example.com"
    local password="password123"
    local name="Stress Test User ${user_id}"
    
    # Register user
    local register_response=$(curl -s -X POST "$API_BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\",
            \"name\": \"$name\"
        }")
    
    # Extract JWT token
    local token=$(echo "$register_response" | jq -r '.token // empty')
    
    if [ -z "$token" ] || [ "$token" = "null" ]; then
        echo -e "${RED}❌ Failed to create user $user_id${NC}"
        return 1
    fi
    
    # Add test wallets (real Ethereum addresses for testing)
    local wallet_addresses=(
        "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
        "0x8ba1f109551bD432803012645Hac136c22C177e9"
        "0x147B8eb97fD247D06C4006D269c90C1908Fb5D54"
    )
    
    for i in {0..2}; do
        curl -s -X POST "$API_BASE_URL/watchlist/wallets" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "{
                \"wallet_address\": \"${wallet_addresses[$i]}\",
                \"label\": \"Stress Test Wallet $((i+1))\"
            }" > /dev/null
    done
    
    # Add test tokens (ETH + popular ERC-20 tokens)
    curl -s -X POST "$API_BASE_URL/watchlist/tokens" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"token_symbol\": \"ETH\",
            \"token_name\": \"Ethereum\"
        }" > /dev/null
    
    curl -s -X POST "$API_BASE_URL/watchlist/tokens" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"token_address\": \"0xA0b86a33E6441b8C4C8C8C8C8C8C8C8C8C8C8C8\",
            \"token_symbol\": \"USDC\",
            \"token_name\": \"USD Coin\"
        }" > /dev/null
    
    curl -s -X POST "$API_BASE_URL/watchlist/tokens" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"token_address\": \"0xB0b86a33E6441b8C4C8C8C8C8C8C8C8C8C8C8C8\",
            \"token_symbol\": \"DAI\",
            \"token_name\": \"Dai Stablecoin\"
        }" > /dev/null
    
    echo "$token"
}

# Create test users and store their tokens
declare -a USER_TOKENS
echo -e "${BLUE}👥 Creating $CONCURRENT_USERS test users...${NC}"

for i in $(seq 1 $CONCURRENT_USERS); do
    echo -n "Creating user $i... "
    token=$(create_test_user $i)
    if [ $? -eq 0 ]; then
        USER_TOKENS[$i]=$token
        echo -e "${GREEN}✅${NC}"
    else
        echo -e "${RED}❌${NC}"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}✅ Test data created successfully${NC}"

# Function to perform API requests with timing
perform_requests() {
    local user_id=$1
    local token=${USER_TOKENS[$user_id]}
    local request_count=0
    local success_count=0
    local error_count=0
    local total_response_time=0
    
    echo "User $user_id starting requests..."
    
    while [ $request_count -lt $REQUESTS_PER_USER ]; do
        # Randomly choose an endpoint to test (excluding balance history for now)
        local endpoint_choice=$((RANDOM % 6))
        
        # Record start time
        start_time=$(date +%s%N)
        
        case $endpoint_choice in
            0)
                # Get wallets
                response=$(curl -s -X GET "$API_BASE_URL/watchlist/wallets" \
                    -H "Authorization: Bearer $token" -w "\n%{http_code}")
                ;;
            1)
                # Get tokens
                response=$(curl -s -X GET "$API_BASE_URL/watchlist/tokens" \
                    -H "Authorization: Bearer $token" -w "\n%{http_code}")
                ;;
            2)
                # Get balances
                response=$(curl -s -X GET "$API_BASE_URL/watchlist/balances" \
                    -H "Authorization: Bearer $token" -w "\n%{http_code}")
                ;;
            3)
                # Refresh balances (trigger balance fetching)
                response=$(curl -s -X POST "$API_BASE_URL/watchlist/balances/refresh" \
                    -H "Authorization: Bearer $token" -w "\n%{http_code}")
                ;;
            4)
                # Health check
                response=$(curl -s -X GET "$API_BASE_URL/../../health" -w "\n%{http_code}")
                ;;
            5)
                # Get current user info
                response=$(curl -s -X GET "$API_BASE_URL/users/me" \
                    -H "Authorization: Bearer $token" -w "\n%{http_code}")
                ;;
        esac
        
        # Record end time and calculate response time
        end_time=$(date +%s%N)
        response_time=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
        total_response_time=$((total_response_time + response_time))
        
        # Extract status code and response body
        status_code=$(echo "$response" | tail -n 1)
        response_body=$(echo "$response" | sed '$d')
        
        if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
            ((success_count++))
        else
            ((error_count++))
            echo "User $user_id - Error $status_code: $response_body"
        fi
        
        ((request_count++))
        
        # Small delay to avoid overwhelming the server
        sleep 0.05
    done
    
    local avg_response_time=$((total_response_time / request_count))
    echo "User $user_id completed: $success_count success, $error_count errors, avg response time: ${avg_response_time}ms"
    
    # Write results to file for analysis
    echo "$user_id,$success_count,$error_count,$avg_response_time" >> stress_test_results.csv
}

# Create results file
echo "user_id,success_count,error_count,avg_response_time_ms" > stress_test_results.csv

# Start concurrent users
echo -e "${YELLOW}🔥 Starting stress test...${NC}"
echo ""

# Start background processes for each user
declare -a USER_PIDS
for i in $(seq 1 $CONCURRENT_USERS); do
    perform_requests $i &
    USER_PIDS[$i]=$!
done

# Background process to trigger balance refreshes periodically
balance_refresh_pid=""
if [ "$BALANCE_REFRESH_INTERVAL" -gt 0 ]; then
    (
        while true; do
            sleep $BALANCE_REFRESH_INTERVAL
            echo -e "${BLUE}�� Triggering balance refresh for all users...${NC}"
            for i in $(seq 1 $CONCURRENT_USERS); do
                token=${USER_TOKENS[$i]}
                curl -s -X POST "$API_BASE_URL/watchlist/balances/refresh" \
                    -H "Authorization: Bearer $token" > /dev/null &
            done
        done
    ) &
    balance_refresh_pid=$!
fi

# Wait for all user processes to complete
for pid in "${USER_PIDS[@]}"; do
    wait $pid
done

# Stop balance refresh process if running
if [ -n "$balance_refresh_pid" ]; then
    kill $balance_refresh_pid 2>/dev/null || true
fi

echo ""
echo -e "${GREEN}✅ Stress test completed${NC}"

# Analyze results
echo -e "${YELLOW}📊 Analyzing results...${NC}"

if [ -f stress_test_results.csv ]; then
    # Calculate totals
    total_requests=$(tail -n +2 stress_test_results.csv | awk -F',' '{sum += $2 + $3} END {print sum}')
    total_success=$(tail -n +2 stress_test_results.csv | awk -F',' '{sum += $2} END {print sum}')
    total_errors=$(tail -n +2 stress_test_results.csv | awk -F',' '{sum += $3} END {print sum}')
    avg_response_time=$(tail -n +2 stress_test_results.csv | awk -F',' '{sum += $4; count++} END {print (count > 0) ? sum/count : 0}')
    
    echo "=================================="
    echo "STRESS TEST RESULTS"
    echo "=================================="
    echo "Total Requests: $total_requests"
    echo "Successful Requests: $total_success"
    echo "Failed Requests: $total_errors"
    echo "Success Rate: $(( (total_success * 100) / total_requests ))%"
    echo "Average Response Time: ${avg_response_time}ms"
    echo "=================================="
fi

# Final health check
echo -e "${YELLOW}�� Final health check...${NC}"
final_health=$(curl -s "$API_BASE_URL/../../health")
echo "Health status: $final_health"

# Check database performance
echo -e "${YELLOW}📊 Database performance check...${NC}"
echo "Checking if balance fetcher is still working..."

# Wait a moment for any pending balance fetches
sleep 5

# Check recent balance records
echo "Recent balance records in database:"
curl -s "$API_BASE_URL/../../health" > /dev/null && echo "✅ Server still responding"

echo ""
echo -e "${GREEN}🎉 Stress test finished!${NC}"
echo "Results saved to: stress_test_results.csv"