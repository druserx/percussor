CC      = go build
TARGET  = percussor
LDFLAGS = -ldflags="-s -w"

all: $(TARGET)

$(TARGET):
	$(CC) $(LDFLAGS) -o $(TARGET) .

windows:
	GOOS=windows GOARCH=amd64 $(CC) $(LDFLAGS) -o $(TARGET).exe .

clean:
	rm -f $(TARGET) $(TARGET).exe

.PHONY: all windows clean
