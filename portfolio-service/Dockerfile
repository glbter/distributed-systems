FROM python:3.11.4-slim-buster

WORKDIR /app

COPY requirements.txt requirements.txt
RUN pip3 install -r requirements.txt

COPY . .

#EXPOSE 8080
#ENV FLASK_APP=main
# CMD [ "python3", "-m" , "flask", "run"]
CMD [ "python3", "async.py"]

