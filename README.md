# Automated RSS service

An RSS feed checking daemon with an HTTP interface for monitoring and subscribing/unsubscribing feeds.

If executed without any arguments it does nothing with the articles it discovers, which makes it kinda pointless! Give it ```-stdout``` on the command line and it will output each article to stdout as a JSON object. Use ```-zmq_pub_address=tcp://*:6543``` to publish items (pub/sub - ARSS binds) to a zmq address. See ```subscriber.py``` for an example consumer.

For full options run it with ```-help``` or see [the code](https://github.com/3ft9/arss/blob/master/arss.go#L238).

## License

"THE BEER-WARE LICENSE" (Revision 42):
<stuart@3ft9.com> wrote this code. As long as you retain this notice you can do whatever you want with this stuff. If we meet some day, and you think this stuff is worth it, you can buy me a ~~beer~~ spiced rum and coke in return.

## Contact

Stuart Dallas<br />
3ft9 Ltd<br />
http://3ft9.com/
