#curl -X POST -H "Content-Type: application/json" -d '{"username":"test","nickname":"TEST","salt":"'$SALT'","password":"'$HASH'"}' http://localhost:8080/register
REFRESH_MSG=$(curl -s -X GET -H "Authorization: bearer $REFRESH_TOKEN" http://localhost:8080/refresh)
echo $REFRESH_MSG | jq

echo ACCESS_TOKEN=$(echo $REFRESH_MSG | jq -r '.access_token')
ACCESS_TOKEN=$(echo $REFRESH_MSG | jq -r '.access_token')
echo REFRESH_TOKEN=$(echo $REFRESH_MSG | jq -r '.refresh_token')
REFRESH_TOKEN=$(echo $REFRESH_MSG | jq -r '.refresh_token')

unset REFRESH_MSG
