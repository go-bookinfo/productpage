FROM alpine:3.17
ADD productpage /app/productpage
ADD index.html /app/index.html
EXPOSE 80
ENTRYPOINT [ "/app/productpage" ]