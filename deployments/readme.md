
## Systemd

### Create user and group

```bash
sudo groupadd --system aienvoy
sudo useradd --system \
    --gid aienvoy\
    --create-home \
    --home-dir /var/lib/aienvoy\
    --shell /usr/sbin/nologin \
    --comment "aienvoy server" \
    aienvoy
```

### Create service file and folder

```bash
sudo mkdir -p /var/lib/aienvoy
sudo ln -s /usr/local/bin/aienvoy /var/lib/aienvoy/aienvoy
sudo chown -R aienvoy:aienvoy /var/lib/aienvoy
```


### Start service

```bash
sudo systemctl daemon-reload
sudo systemctl enable aienvoy
sudo systemctl start aienvoy
sudo systemctl status aienvoy
```
