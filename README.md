Use evtest utility or cat /proc/bus/input/devices to determine which input event device number is your keyboard
```bash
docker run --privileged -v /dev/:/dev/ -p 9121:9121 -e INPUT_DEVICE=/dev/input/event4 -e PORT=9121 tombokombo/keypress-exporter
```
