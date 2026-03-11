#curl -X POST -H "Content-Type: application/json" -d '{"username":"test","nickname":"TEST","salt":"'$SALT'","password":"'$HASH'"}' http://localhost:8080/register
curl -s -X GET -H "Authorization: bearer $ACCESS_TOKEN" http://localhost:8080/admin/profile
