#curl -X POST -H "Content-Type: application/json" -d '{"username":"test","nickname":"TEST","salt":"'$SALT'","password":"'$HASH'"}' http://localhost:8080/register
LOGOUT_MSG=$(curl -s -X GET -H "Authorization: bearer $ACCESS_TOKEN" http://localhost:8080/logout)
echo $LOGOUT_MSG | jq

unset LOGOUT_MSG
