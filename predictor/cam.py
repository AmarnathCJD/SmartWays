import cv2
import cvzone
import math
import numpy as np
from ultralytics import YOLO
from sort import *
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt

analyse = input("Analysis or Prediction? (a/p): ")

if analyse == "p":
    print("\n🚦 Analysing the Video 🚦\n")
    video_path = "vids/jn-vid.mp4"
    cap = cv2.VideoCapture(video_path)
    model = YOLO("models/yolo11n.pt")
    # model.to("cuda")

    input("Press Enter to Start the Video Analysis...")

    classnames = ["car", "truck", "bus"]

    road_zoneA = np.array(
        [[308, 789], [711, 807], [694, 392], [415, 392], [309, 790]], np.int32
    )
    road_zoneB = np.array(
        [[727, 797], [1123, 812], [1001, 516], [741, 525], [730, 795]], np.int32
    )
    road_zoneC = np.array(
        [[1116, 701], [1533, 581], [1236, 367], [1009, 442], [1122, 698]], np.int32
    )

    zoneA_Line = np.array([road_zoneA[0], road_zoneA[1]]).reshape(-1)
    zoneB_Line = np.array([road_zoneB[0], road_zoneB[1]]).reshape(-1)
    zoneC_Line = np.array([road_zoneC[0], road_zoneC[1]]).reshape(-1)

    tracker = Sort()
    zoneAcounter = []
    zoneBcounter = []
    zoneCcounter = []

    vehicle_counts = []
    frame_times = []

    frame_count = 0

    while True:
        ret, frame = cap.read()
        try:
            frame = cv2.resize(frame, (1920, 1080))
        except:
            break
        results = model(frame)
        current_detections = np.empty([0, 5])

        for info in results:
            parameters = info.boxes
            for box in parameters:
                x1, y1, x2, y2 = box.xyxy[0]
                x1, y1, x2, y2 = int(x1), int(y1), int(x2), int(y2)
                w, h = x2 - x1, y2 - y1
                confidence = box.conf[0]
                class_detect = box.cls[0]
                class_detect = int(class_detect)
                try:
                    class_detect = classnames[class_detect]
                except:
                    class_detect = "unknown"
                conf = math.ceil(confidence * 100)
                cvzone.putTextRect(
                    frame, f"{class_detect}", [x1 + 8, y1 - 12], thickness=2, scale=1
                )
                cv2.rectangle(frame, (x1, y1), (x2, y2), (0, 255, 0), 2)

                if (
                    class_detect == "car"
                    or class_detect == "truck"
                    or class_detect == "bus"
                    and conf > 60
                ):
                    detections = np.array([x1, y1, x2, y2, conf])
                    current_detections = np.vstack([current_detections, detections])

        # cv2.polylines(frame, [road_zoneA], isClosed=False, color=(0, 0, 255), thickness=8)
        # cv2.polylines(frame, [road_zoneB], isClosed=False, color=(0, 255, 255), thickness=8)
        # cv2.polylines(frame, [road_zoneC], isClosed=False, color=(255, 0, 0), thickness=8)

        track_results = tracker.update(current_detections)
        for result in track_results:
            x1, y1, x2, y2, id = result
            x1, y1, x2, y2, id = int(x1), int(y1), int(x2), int(y2), int(id)
            w, h = x2 - x1, y2 - y1
            cx, cy = x1 + w // 2, y1 + h // 2 - 40

            if (
                zoneA_Line[0] < cx < zoneA_Line[2]
                and zoneA_Line[1] - 20 < cy < zoneA_Line[1] + 20
            ):
                if zoneAcounter.count(id) == 0:
                    zoneAcounter.append(id)

            if (
                zoneB_Line[0] < cx < zoneB_Line[2]
                and zoneB_Line[1] - 20 < cy < zoneB_Line[1] + 20
            ):
                if zoneBcounter.count(id) == 0:
                    zoneBcounter.append(id)

            if (
                zoneC_Line[0] < cx < zoneC_Line[2]
                and zoneC_Line[1] - 20 < cy < zoneC_Line[1] + 20
            ):
                if zoneCcounter.count(id) == 0:
                    zoneCcounter.append(id)

            cv2.line(frame, (cx, cy), (cx, cy + 40), (0, 255, 0), 2)
            cv2.circle(frame, (cx, cy), 5, (0, 0, 255), cv2.FILLED)
            cv2.rectangle(frame, (x1, y1), (x2, y2), (255, 0, 0), 2)
            cvzone.putTextRect(
                frame,
                f"SmartWays - CAM Analyser",
                [10, 50],
                scale=2,
                thickness=2,
                colorT=(255, 255, 255),
                colorR=(255, 0, 255),
                font=cv2.FONT_HERSHEY_PLAIN,
            )
            cvzone.putTextRect(
                frame,
                f"Vehicle Passed: {len(zoneAcounter) + len(zoneBcounter) + len(zoneCcounter)}",
                [10, 100],
                scale=2,
                thickness=2,
                colorT=(255, 255, 255),
                colorR=(255, 0, 255),
                font=cv2.FONT_HERSHEY_PLAIN,
            )
            cvzone.putTextRect(
                frame,
                f"Zone A: {len(zoneAcounter)}",
                [10, 150],
                scale=2,
                thickness=2,
            )
            cvzone.putTextRect(
                frame,
                f"Zone B: {len(zoneBcounter)}",
                [10, 200],
                scale=2,
                thickness=2,
            )
            cvzone.putTextRect(
                frame,
                f"Zone C: {len(zoneCcounter)}",
                [10, 250],
                scale=2,
                thickness=2,
            )
            cvzone.putTextRect(
                frame,
                f"TVehicles on Road: {len(parameters)}",
                [10, 300],
                scale=2,
                thickness=2,
            )
            cvzone.putTextRect(
                frame,
                f"ID: {id}",
                [cx, cy],
                scale=1,
                thickness=2,
            )
            cv2.rectangle(frame, (x1, y1), (x2, y2), (255, 0, 0), 2)
            cv2.circle(frame, (cx, cy), 5, (0, 0, 255), cv2.FILLED)

        cv2.imshow("SmartWays - CAM Analyser", frame)
        frame_count += 1
        frame_times.append(frame_count / 30)
        vehicle_counts.append(len(parameters))
        if cv2.waitKey(1) & 0xFF == ord("q"):
            break

    cap.release()
    cv2.destroyAllWindows()

    df = pd.DataFrame({"Time (s)": frame_times, "Vehicle Count": vehicle_counts})
    df.to_csv("results/vehicle_counts.csv", index=False)

print("\n🚦 Traffic Analysis Report 🚦\n\n")
df = pd.read_csv("results/vehicle_counts.csv")
plt.figure(figsize=(12, 6))
sns.lineplot(
    x=df["Time (s)"], y=df["Vehicle Count"], marker="o", linestyle="-", color="b"
)

plt.title("Traffic Flow Over Time", fontsize=14)
plt.xlabel("Time (seconds)", fontsize=12)
plt.ylabel("Number of Vehicles", fontsize=12)
plt.grid(True)
plt.savefig("results/traffic_analysis.png")

mean_vehicles = df["Vehicle Count"].mean()
max_vehicles = df["Vehicle Count"].max()
min_vehicles = df["Vehicle Count"].min()

print(f"✅ Average Vehicle Count: {mean_vehicles:.2f}")
print(f"🚗 Peak Traffic (Max Vehicles): {max_vehicles}")
print(f"🟢 Least Traffic (Min Vehicles): {min_vehicles}")

peak_times = df[df["Vehicle Count"] == max_vehicles]["Time (s)"].values
print(f"⚠️ Peak Traffic Time(s): {peak_times}")

# Define congestion threshold (e.g., > 80% of max vehicles)
congestion_threshold = max_vehicles * 0.8
congested_times = df[df["Vehicle Count"] >= congestion_threshold]

if not congested_times.empty:
    print("🚦 High Congestion Detected at:")
    print(congested_times)
else:
    print("✅ No significant congestion detected.")
