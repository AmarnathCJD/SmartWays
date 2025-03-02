import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from sklearn.preprocessing import MinMaxScaler

from sklearn.model_selection import train_test_split

df = pd.read_csv("vehicle_counts.csv")

scaler = MinMaxScaler()
df["Vehicle Count"] = scaler.fit_transform(df[["Vehicle Count"]])

df.set_index("Time (s)", inplace=True)

# Plot the preprocessed data
plt.figure(figsize=(12, 6))
plt.plot(df.index, df["Vehicle Count"], label="Normalized Vehicle Count", color="b")
plt.title("Normalized Traffic Data for LSTM")
plt.xlabel("Time (seconds)")
plt.ylabel("Normalized Vehicle Count")
plt.legend()
plt.grid(True)
plt.savefig("results/preprocessed_data.png")

print("✅ Data Preprocessed for LSTM!")


def create_sequences(data, seq_length=30):
    X, Y = [], []
    for i in range(len(data) - seq_length):
        X.append(data[i : i + seq_length])
        Y.append(data[i + seq_length])
    return np.array(X), np.array(Y)


# Define sequence length (e.g., last 30 seconds)
SEQ_LENGTH = 30


X, Y = create_sequences(df["Vehicle Count"].values, SEQ_LENGTH)
X = X.reshape(X.shape[0], X.shape[1], 1)

print(f"✅ Data Shape for LSTM -> X: {X.shape}, Y: {Y.shape}")
X_train, X_test, Y_train, Y_test = train_test_split(X, Y, test_size=0.2, shuffle=False)

print(f"✅ Training Data: {X_train.shape}, Testing Data: {X_test.shape}")

import tensorflow as tf
from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import (
    LSTM,
    Bidirectional,
    Dense,
    Dropout,
    BatchNormalization,
)

model = Sequential(
    [
        Bidirectional(
            LSTM(128, return_sequences=True, input_shape=(X_train.shape[1], 1))
        ),
        BatchNormalization(),
        Dropout(0.3),
        LSTM(64, return_sequences=True),
        BatchNormalization(),
        Dropout(0.3),
        LSTM(32, return_sequences=False),
        BatchNormalization(),
        Dropout(0.3),
        Dense(32, activation="relu"),
        Dense(1),
    ]
)

model.compile(
    optimizer=tf.keras.optimizers.Adam(learning_rate=0.0005),
    loss="mse",
    metrics=["mae"],
)
model.summary()

from tensorflow.keras.callbacks import EarlyStopping, ReduceLROnPlateau

early_stop = EarlyStopping(monitor="val_loss", patience=10, restore_best_weights=True)
reduce_lr = ReduceLROnPlateau(monitor="val_loss", factor=0.5, patience=5, min_lr=1e-6)

history = model.fit(
    X_train,
    Y_train,
    epochs=100,
    batch_size=16,
    validation_data=(X_test, Y_test),
    callbacks=[early_stop, reduce_lr],
    verbose=1,
)

import matplotlib.pyplot as plt

plt.plot(history.history["loss"], label="Train Loss")
plt.plot(history.history["val_loss"], label="Validation Loss")
plt.title("Fine-Tuned LSTM Model Training Loss")
plt.xlabel("Epochs")
plt.ylabel("Loss")
plt.legend()
plt.savefig("results/training_loss.png")

model.save("results/trained_model.keras")

predictions = model.predict(X_test)

predictions = scaler.inverse_transform(predictions)
Y_test_actual = scaler.inverse_transform(Y_test.reshape(-1, 1))

plt.figure(figsize=(12, 6))
plt.plot(Y_test_actual, label="Actual Vehicle Count", color="blue")
plt.plot(predictions, label="Predicted Vehicle Count", color="red", linestyle="dashed")
plt.title("Fine-Tuned LSTM Traffic Prediction")
plt.xlabel("Time")
plt.ylabel("Vehicle Count")
plt.legend()
plt.grid(True)
plt.savefig("results/traffic_prediction.png")
