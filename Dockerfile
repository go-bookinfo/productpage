FROM alpine:3.17
ADD productpage /app/productpage
EXPOSE 8000
ENTRYPOINT [ "/app/productpage" ]