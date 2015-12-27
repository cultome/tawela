import argparse
import datetime
import imutils
import time
import cv2

camera = cv2.VideoCapture("rtsp://192.168.1.128:554/11")
firstFrame = None
while True:
  (grabbed, frame) = camera.read()
  print("[*] Taking one frame")
  if not grabbed:
    break

  frame = imutils.resize(frame, width=500)
  gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
  gray = cv2.GaussianBlur(gray, (21, 21), 0)

  if firstFrame is None:
    firstFrame = gray
    cv2.imwrite("first.jpg", frame)
    time.sleep(5.0)
    continue

  frameDelta = cv2.absdiff(firstFrame, gray)
  thresh = cv2.threshold(frameDelta, 25, 255, cv2.THRESH_BINARY)[1]

  thresh = cv2.dilate(thresh, None, iterations=2)
  (cnts, _) = cv2.findContours(thresh.copy(), cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

  for c in cnts:
    if cv2.contourArea(c) < 500:
      continue
    (x, y, w, h) = cv2.boundingRect(c)
    cv2.rectangle(frame, (x, y), (x + w, y + h), (0, 255, 0), 2)

  cv2.imwrite("detection.jpg", frame)
  cv2.imwrite("delta.jpg", frameDelta)
  cv2.imwrite("tresh.jpg", thresh)
  break

camera.release()
cv2.destroyAllWindows()
