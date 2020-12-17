FROM postgres:12.2

COPY scripts/pg/init-db.sh /

ENTRYPOINT [ "/init-db.sh" ]
