ssh-keygen -t rsa -b 4096 -m PEM -E SHA512 -f .keys/jwtRS512.key -N "" && openssl rsa -in .keys/jwtRS512.key -pubout -outform PEM -out .keys/jwtRS512.key.pub
