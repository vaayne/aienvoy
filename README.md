
## Admin

user: admin@admin.com
pass: adminadmin

test_user@admin.com
test_usertest_user

## API

### Chat

PATH: `/v1/chat/completions`

PAYLOAD: {
    "model": "",
    "messages": [{}]
}



### Embedding

PATH: `/v1/embeddings`

PAYLOAD: {
    "model": "text-embedding-ada-002",
    "input": ""
}


## Deployment

### Systemd

1. add user to limit privileges
    `sudo useradd -s /sbin/nologin -M odb`
2. Copy systemd unit service file
   ```bash
   sudo cp ./deployments/odb.service /etc/systemd/system/
   sudo chmod 755 /lib/systemd/system/odb.service
   sudo systemctl daemon-reload
   sudo systemctl enable odb.service
   sudo systemctl start odb
   # tail logs
   sudo journalctl -f -u odb
   ```
