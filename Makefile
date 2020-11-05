all: deliverable-receiver.linux64

deliverable-receiver.linux64:
	GOOS=linux GOARCH=amd64 go build -o deliverable-receiver.linux64

docker: deliverable-receiver.linux64
	docker build . -t guardianmultimedia/deliverable-receiver:DEV