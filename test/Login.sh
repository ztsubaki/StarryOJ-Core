NAME="test"
SALT=$(pwgen -s 16 1)
PASSWD="test"
HASH=$(echo -n "$PASSWD$SALT" | sha256sum | awk '{print $1}')

PRELOGIN_MSG=$(curl -s http://localhost:8080/prelogin/$NAME | jq)
echo $PRELOGIN_MSG | jq
SALT=$(echo $PRELOGIN_MSG | jq -r '.data.salt')
LOGINSALT=$(echo $PRELOGIN_MSG | jq -r '.data.login_salt')

PASSWORD=$(echo -n "$PASSWD$SALT" | sha256sum | awk '{print $1}')
LOGINHASH=$(echo -n "$PASSWORD$LOGINSALT" | sha256sum | awk '{print $1}')

#curl -X POST -H "Content-Type: application/json" -d '{"username":"test","nickname":"TEST","salt":"'$SALT'","password":"'$HASH'"}' http://localhost:8080/register
LOGIN_MSG=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"test","password":"'$LOGINHASH'"}' http://localhost:8080/login)

echo $LOGIN_MSG | jq

echo ACCESS_TOKEN=$(echo $LOGIN_MSG | jq -r '.data.access_token')
ACCESS_TOKEN=$(echo $LOGIN_MSG | jq -r '.data.access_token')
echo REFRESH_TOKEN=$(echo $LOGIN_MSG | jq -r '.data.refresh_token')
REFRESH_TOKEN=$(echo $LOGIN_MSG | jq -r '.data.refresh_token')

unset USERNAME
unset SALT
unset HASH
unset LOGINSALT
unset PASSWORD
unset LOGINHASH
unset PRELOGIN_MSG
unset LOGIN_MSG

