import pika
import msgpack

import concurrent.futures
from threading import current_thread, get_ident, get_native_id

import base64
from io import BytesIO

import torchaudio

import boto3

connection = pika.BlockingConnection(pika.URLParameters("amqp://guest:leo869636@localhost"))
channel = connection.channel()

def processAction(action):
    tag, act = action

    _, encoded = act['data'].split("base64,", 1)
    data = base64.b64decode(encoded)
    with BytesIO(data) as audioFile:
        metadata = torchaudio.info(audioFile)
        print(metadata)
    channel.basic_ack(tag)


with concurrent.futures.ThreadPoolExecutor(max_workers=4) as exec:
    for method_frame, properties, body in channel.consume("actions"):
        # Display the message parts and acknowledge the message
        # print(method_frame, properties)
        decoded = msgpack.unpackb(body, use_list=False, raw=False)

        exec.map(processAction, [(method_frame.delivery_tag, decoded)])
        

# Cancel the consumer and return any pending messages
requeued_messages = channel.cancel()
print('Requeued %i messages' % requeued_messages)

connection.close()