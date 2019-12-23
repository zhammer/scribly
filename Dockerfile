FROM python:3.7

COPY . ./project
WORKDIR /project

RUN pip install --no-cache-dir -r requirements.txt
