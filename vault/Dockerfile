FROM vault:latest
ADD setup.sh /root
ADD init-vault-container.sh /root
ENTRYPOINT ["vault", "server", "-dev"]
