FROM vault:latest
ADD setup.sh /root
ADD init-vault-container.sh /root
ENTRYPOINT ["vault", "server", "-dev", "-dev-listen-address=0.0.0.0:8200"]
