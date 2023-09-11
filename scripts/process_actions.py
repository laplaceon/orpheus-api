import os

import pika
import msgpack
from dotenv import load_dotenv
import concurrent.futures

import base64
from io import BytesIO

from pydub import AudioSegment

import torchaudio

import boto3
from botocore.config import Config

import uuid

from db import Database

load_dotenv()

db = Database(os.getenv("DSN"))

bucket_name = "tuneforge"

connection = pika.BlockingConnection(pika.URLParameters(os.getenv('RABBIT_CONN')))
channel = connection.channel()

my_config = Config(
    signature_version='v4',
)

s3 = boto3.resource('s3',
    endpoint_url = os.getenv('R2_ENDPOINT'),
    aws_access_key_id = os.getenv('R2_ACCESS_KEY_ID'),
    aws_secret_access_key = os.getenv('R2_ACCESS_KEY_SECRET'),
    config = my_config
)

bucket = s3.Bucket(bucket_name)

def processAction(action):
    tag, act = action

    exp_hours, status = db.get_history_item(act["history_id"])

    if status == 0:
        _, encoded = act['data'].split("base64,", 1)
        data = base64.b64decode(encoded)
        with BytesIO(data) as audioFile:
            data, rate = torchaudio.load(audioFile)
        print(data, rate)

        with BytesIO() as bytes:
            unique_filename = str(uuid.uuid4()) + ".mp3"
            torchaudio.save(bytes, data, rate, format="wav")
            bytes.seek(0)

            audioFile = AudioSegment.from_file(bytes, frame_rate=rate, format="wav")
            bytes.seek(0)
            bytes.truncate(0)

            params = {"format": "mp3", "bitrate": "320k"}

            audioFile.export(bytes, **params)
            bytes.seek(0)

            bucket.upload_fileobj(bytes, f"gen/{unique_filename}")

            client = s3.meta.client

            url = client.generate_presigned_url('get_object', Params={'Bucket': bucket_name, 'Key': f"gen/{unique_filename}"}, ExpiresIn=exp_hours * 60 * 60)
            print(url)

    # channel.basic_ack(tag)


with concurrent.futures.ThreadPoolExecutor(max_workers=4) as exec:
    for method_frame, properties, body in channel.consume("actions"):
        # Display the message parts and acknowledge the message
        # print(method_frame, properties)
        decoded = msgpack.unpackb(body, use_list=False, raw=False)

        exec.map(processAction, [(method_frame.delivery_tag, decoded)])
        

# for method_frame, properties, body in channel.consume("actions"):
#     decoded = msgpack.unpackb(body, use_list=False, raw=False)

#     processAction((method_frame.delivery_tag, decoded))

# Cancel the consumer and return any pending messages
requeued_messages = channel.cancel()
print('Requeued %i messages' % requeued_messages)

connection.close()
db.quit()