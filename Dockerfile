FROM python:3.7

COPY . ./project
WORKDIR /project

RUN pip install -r requirements.txt
