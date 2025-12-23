#!/bin/bash

# ==============================================================================
# Налаштування
# ==============================================================================
API_URL="http://localhost:8000/integrations/transaction-indexing-svc"
BTC_CLI="docker exec bitcoin-node bitcoin-cli -regtest -rpcuser=user -rpcpassword=pass"
USER="user_$(date +%s)"
PASS="password123"

echo "--- 1. Реєстрація ---"
curl -s -X POST "$API_URL/register" \
     -H "Content-Type: application/json" \
     -d "{\"username\":\"$USER\", \"password\":\"$PASS\"}" | jq .

echo -e "\n--- 2. Авторизація ---"
# Отримуємо токен (сценарій передбачає, що він повертається у полі "token")
TOKEN=$(curl -s -X POST "$API_URL/login" \
     -H "Content-Type: application/json" \
     -d "{\"username\":\"$USER\", \"password\":\"$PASS\"}" | jq -r .token)

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Помилка: Не вдалося отримати JWT токен."
    exit 1
fi
echo "JWT отримано."

echo -e "\n--- 3. Додавання адреси ---"
# Генеруємо нову адресу в Bitcoin Core
TRACKED_ADDR=$($BTC_CLI getnewaddress)
echo "Згенерована адреса: $TRACKED_ADDR"

curl -s -X POST "$API_URL/addresses" \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d "{\"address\":\"$TRACKED_ADDR\"}" | jq .

echo -e "\n--- 4. Генерація монет та транзакція (BTC Core) ---"
# Майнимо 101 блок (щоб з'явилися кошти)
$BTC_CLI generatetoaddress 101 $($BTC_CLI getnewaddress) > /dev/null
# Надсилаємо 1.5 BTC на нашу адресу
TXID=$($BTC_CLI sendtoaddress "$TRACKED_ADDR" 1.5)
echo "Транзакція надіслана: $TXID"
# Підтверджуємо транзакцію (майнимо ще 1 блок)
$BTC_CLI generatetoaddress 1 $($BTC_CLI getnewaddress) > /dev/null

echo "Очікуємо 10 секунд на індексацію..."
sleep 10

echo -e "\n--- 5. Перевірка балансу ---"
curl -s -X GET "$API_URL/addresses/$TRACKED_ADDR/balance" \
     -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n--- 6. Перевірка UTXO ---"
curl -s -X GET "$API_URL/addresses/$TRACKED_ADDR/utxos" \
     -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n--- 7. Перевірка історії транзакцій ---"
curl -s -X GET "$API_URL/addresses/$TRACKED_ADDR/txs" \
     -H "Authorization: Bearer $TOKEN" | jq .