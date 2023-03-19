FROM alpine:3.17
ADD productpage /app/productpage
EXPOSE 80
ENTRYPOINT [ "/app/productpage" ]