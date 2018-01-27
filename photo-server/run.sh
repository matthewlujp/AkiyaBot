#!/bin/sh

echo 'Create ssh tunnel to a proxy server and run a server to listen for photo request.

Before running the server, following environment variables are required.
--------------------------------------------------------------------
PORT="port number which the photo server listens on. default to 4050"
PROXY_SERVER_HOST="address to a proxy server"
PROXY_SERVER_USER="user to which ssh connection is created"
SSH_LOG_FILE="file to output ssh tunnel logs. default to /var/log/ssh_tunnel"
SERVER_LOG_FILE="file to output photo server logs. default to /var/log/photo_server"
SERVER_BIN="exec file path of photo server"
RSA_SECRET_KEY="rsa secret key file path to access a proxy server"
--------------------------------------------------------------------
'

if [ -z $PORT ]; then
  PORT=4050;
fi
if [ -z $SSH_LOG_FILE ]; then
  SSH_LOG_FILE=/var/log/ssh_tunnel;
fi
if [ -z $SERVER_LOG_FILE ]; then
  SERVER_LOG_FILE=/var/log/photo_server;
fi

if [ -z $PROXY_SERVER_HOST ]; then
  echo "\$PROXY_SERVER_HOST is not set"
  exit
fi
if [ -z $PROXY_SERVER_USER ]; then
  echo "\$PROXY_SERVER_USER is not set"
  exit
fi
if [ -z $SERVER_BIN ]; then
  echo "\$SERVER_BIN is not set"
  exit
fi
if [ -z $RSA_SECRET_KEY ]; then
  echo "\$RSA_SECRET_KEY is not set"
  exit
fi

ssh_tunnel() {
  while :
  do
    ssh -i $RSA_SECRET_KEY -R $1:localhost:$1 $2 -N -o ServerAliveInterval=20 -o ServerAliveCountMax=3 >> $3 2>&1
  done
}

ssh_tunnel $PORT $PROXY_SERVER_USER@$PROXY_SERVER_HOST $SSH_LOG_FILE &
$SERVER_BIN -p $PORT >> $SERVER_LOG_FILE 2>&1 &
