ARG ALLINONE=default
FROM $ALLINONE as allinone

FROM gcr.io/distroless/base-debian10:debug

COPY --from=allinone /usr/local/bin/frm /usr/local/bin/frm
