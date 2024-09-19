FROM golang:1.19

WORKDIR /opt/hyperledger/chaincode/hospital

COPY . .

RUN go mod download
RUN go build -o hospital

CMD ["./hospital"]
