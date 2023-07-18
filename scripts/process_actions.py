import pika
import msgpack

import concurrent.futures

import base64
from io import BytesIO

import torchaudio

import boto3
from botocore.config import Config

import uuid

connection = pika.BlockingConnection(pika.URLParameters("amqp://guest:leo869636@localhost"))
channel = connection.channel()

my_config = Config(
    signature_version='v4',
)

s3 = boto3.resource('s3',
    endpoint_url = 'https://4a2ec92e72b8d8d4cbc6299d60a7fc78.r2.cloudflarestorage.com',
    aws_access_key_id = 'bcd9890ef1859e0eabfea3f33b7ebc76',
    aws_secret_access_key = 'c57960ac2f1339299cc7b83d82c1f65aed1ebac3527f6500ee49a3a2852205a5',
    config = my_config
)

bucket = s3.Bucket('orpheus')

def processAction(action):
    tag, act = action

    _, encoded = act['data'].split("base64,", 1)
    data = base64.b64decode(encoded)
    with BytesIO(data) as audioFile:
        data, rate = torchaudio.load(audioFile)
    print(data)

    with BytesIO() as audioFile:
        unique_filename = str(uuid.uuid4())
        torchaudio.save(audioFile, data, rate, format="wav")
        audioFile.seek(0)
        bucket.upload_fileobj(audioFile, f"{unique_filename}.wav")

        client = s3.meta.client

        url = client.generate_presigned_url('get_object', Params={'Bucket': 'orpheus', 'Key': f"{unique_filename}.wav"}, ExpiresIn=3600)
        print(url)

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