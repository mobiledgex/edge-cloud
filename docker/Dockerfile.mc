ARG ALLINONE=default
FROM $ALLINONE as allinone

FROM gcr.io/distroless/base-debian10:debug

COPY --from=allinone /usr/local/bin/mc /usr/local/bin/mc
COPY --from=allinone /MobiledgeX_Logo.png /MobiledgeX_Logo.png
