USERNAME="test"
SALT=$(pwgen -s 16 1)
PASSWD="test"
HASH=$(echo -n "$PASSWD$SALT" | sha256sum | awk '{print $1}')

curl http://localhost:8080/prelogin/$USERNAME
