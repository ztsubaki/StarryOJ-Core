SALT=$(pwgen -s 16 1)
PASSWD="test"
HASH=$(echo -n "$PASSWD$SALT" | sha256sum | awk '{print $1}')

echo "Salt: $SALT"
echo "Hash: $HASH"

curl -X POST -H "Content-Type: application/json" -d '{"username":"test","nickname":"TEST","salt":"'$SALT'","password":"'$HASH'"}' http://localhost:8080/register
