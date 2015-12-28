#!/usr/bin/env python

import imutils
import cv2

def getCameraMove(limitPerc):
    base, screenH, screenW = prepareForDiff("base.jpg")
    diff,_,_ = prepareForDiff("diff.jpg")

    cnts = detectContours(base, diff)
    c = findMax(cnts)

    if c is not None:
        (x, y, w, h) = cv2.boundingRect(c)
        saveImage(diff, c)
        return calculateCameraMove((x, y), w,  h, screenW, screenH, limitPerc)

def calculateCameraMove(objCorner, objW, objH, screenW, screenH, limitPerc):
    center = ( objW / 2 + objCorner[0], objH / 2 + objCorner[1] )

    limitSizeW = screenW * limitPerc
    limitSizeH = screenH * limitPerc

    #print("Corner: ", objCorner)
    #print("Center: ", center)
    #print("Object: ", objW, objH)
    #print("Screen: ", screenW, screenH)
    #print("Limits: ", limitSizeW, limitSizeH)

    move = ""

    if center[0] <= limitSizeW and center[1] <= limitSizeH:
        move = "UpLeft"
    elif center[0] <= limitSizeW and center[1] >= screenW - limitSizeH:
        move = "DownLeft"
    elif center[0] >= screenW - limitSizeW and center[1] <= limitSizeH:
        move = "UpRight"
    elif center[0] >= screenW - limitSizeW and center[1] >= screenH - limitSizeH:
        move = "DownRight"
    elif center[0] <= limitSizeW:
        move = "Left"
    elif center[0] >= screenW - limitSizeW:
        move = "Right"
    elif center[1] <= limitSizeH:
        move = "Up"
    elif center[1] >= screenH - limitSizeH:
        move = "Down"

    return move

def saveImage(frame, c):
    (x, y, w, h) = cv2.boundingRect(c)
    cv2.rectangle(frame, (x, y), (x + w, y + h), (0, 255, 0), 2)
    cv2.imwrite("detection.jpg", frame)

def findMax(cnts):
    max = 0
    contour = None

    for c in cnts:
        area = cv2.contourArea(c)
        if area < 500:
            continue
        elif area > max:
            max = area
            contour = c

    return contour

def prepareForDiff(imgPath):
    frame = cv2.imread(imgPath)
    #frame = imutils.resize(frame, width=500)
    gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
    gray = cv2.GaussianBlur(gray, (21, 21), 0)
    w, h = gray.shape
    return gray, w, h

def detectContours(base, diff):
    frameDelta = cv2.absdiff(base, diff)
    thresh = cv2.threshold(frameDelta, 25, 255, cv2.THRESH_BINARY)[1]

    thresh = cv2.dilate(thresh, None, iterations=2)
    (cnts, _) = cv2.findContours(thresh.copy(), cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    return cnts

print(getCameraMove(0.1))
