FROM python:3.9

COPY . ./project
WORKDIR /project

RUN pip install --no-cache-dir -r requirements.txt
