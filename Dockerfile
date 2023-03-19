FROM alpine:3.17
WORKDIR /app
ADD productpage productpage
ADD index.html index.html
EXPOSE 80
ENTRYPOINT [ "/app/productpage" ]