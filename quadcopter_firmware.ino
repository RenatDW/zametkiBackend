#include <Wire.h>
#include <MPU6050.h>

// MPU-6050 объект (адрес по умолчанию 0x68)
MPU6050 mpu(0x68);

// Пины GPIO для ESC
const int ESC_M1 = 25;  // Передний левый
const int ESC_M2 = 26;  // Передний правый
const int ESC_M3 = 27;  // Задний правый
const int ESC_M4 = 14;  // Задний левый

// PWM параметры
const int PWM_FREQ = 50;    // 50 Hz для ESC
const int PWM_RESOLUTION = 8; // 8-bit (0-255)

// Переменные для IMU
float accelX = 0, accelY = 0, accelZ = 0;
float gyroX = 0, gyroY = 0, gyroZ = 0;
float angleX = 0, angleY = 0, angleZ = 0;

// PID параметры
float kpRoll = 1.5, kiRoll = 0.05, kdRoll = 0.8;
float kpPitch = 1.5, kiPitch = 0.05, kdPitch = 0.8;
float kpYaw = 2.0, kiYaw = 0.05, kdYaw = 1.0;

float errorRoll = 0, errorPitch = 0, errorYaw = 0;
float integralRoll = 0, integralPitch = 0, integralYaw = 0;
float prevErrorRoll = 0, prevErrorPitch = 0, prevErrorYaw = 0;
float pidRoll = 0, pidPitch = 0, pidYaw = 0;

uint8_t throttle = 125;
unsigned long lastTime = 0;
float dt = 0.01;

float targetRoll = 0, targetPitch = 0, targetYaw = 0;

void setup() {
  Serial.begin(115200);
  delay(1000);

  Serial.println("Quadcopter Init...");

  // Инициализация I2C
  Wire.begin(21, 22); // SDA=21, SCL=22
  Wire.setClock(400000);
  delay(100);

  // Инициализация MPU-6050
  mpu.initialize();
  if (!mpu.testConnection()) {
    Serial.println("MPU6050 connection failed!");
    while (1);
  }

  Serial.println("MPU6050 initialized");

  // Настройка PWM для ESC (ESP32 ledcAttach API)
  ledcAttach(ESC_M1, PWM_FREQ, PWM_RESOLUTION);
  ledcAttach(ESC_M2, PWM_FREQ, PWM_RESOLUTION);
  ledcAttach(ESC_M3, PWM_FREQ, PWM_RESOLUTION);
  ledcAttach(ESC_M4, PWM_FREQ, PWM_RESOLUTION);

  Serial.println("ESC PWM configured");

  armESC();
  lastTime = millis();
}

void loop() {
  unsigned long currentTime = millis();
  dt = (currentTime - lastTime) / 1000.0;
  lastTime = currentTime;

  if (dt > 0.1) dt = 0.01;

  // Чтение данных с MPU-6050
  int16_t ax, ay, az, gx, gy, gz;
  mpu.getMotion6(&ax, &ay, &az, &gx, &gy, &gz);

  // Нормализация ускорения (обычно ±16g = 32768)
  accelX = ax / 2048.0;
  accelY = ay / 2048.0;
  accelZ = az / 2048.0;

  // Нормализация гироскопа (±2000 dps = 16.4)
  gyroX = gx / 16.4;
  gyroY = gy / 16.4;
  gyroZ = gz / 16.4;

  // Интегрирование гироскопа
  angleX += gyroX * dt;
  angleY += gyroY * dt;
  angleZ += gyroZ * dt;

  // Complementary filter: ускорение корректирует дрейф гироскопа
  float accelAngleX = atan2(accelY, accelZ) * 180 / PI;
  float accelAngleY = atan2(-accelX, sqrt(accelY * accelY + accelZ * accelZ)) * 180 / PI;

  angleX = angleX * 0.98 + accelAngleX * 0.02;
  angleY = angleY * 0.98 + accelAngleY * 0.02;

  // PID контроль
  pidRoll = computePID(angleX, targetRoll, kpRoll, kiRoll, kdRoll,
                       errorRoll, integralRoll, prevErrorRoll, dt);
  pidPitch = computePID(angleY, targetPitch, kpPitch, kiPitch, kdPitch,
                        errorPitch, integralPitch, prevErrorPitch, dt);
  pidYaw = computePID(angleZ, targetYaw, kpYaw, kiYaw, kdYaw,
                      errorYaw, integralYaw, prevErrorYaw, dt);

  // Распределение тяги (Quadcopter X config)
  float motor1 = throttle + pidPitch + pidRoll + pidYaw;
  float motor2 = throttle + pidPitch - pidRoll - pidYaw;
  float motor3 = throttle - pidPitch - pidRoll + pidYaw;
  float motor4 = throttle - pidPitch + pidRoll - pidYaw;

  // Ограничение и отправка
  motor1 = constrain(motor1, 125, 255);
  motor2 = constrain(motor2, 125, 255);
  motor3 = constrain(motor3, 125, 255);
  motor4 = constrain(motor4, 125, 255);

  ledcWrite(ESC_M1, (uint8_t)motor1);
  ledcWrite(ESC_M2, (uint8_t)motor2);
  ledcWrite(ESC_M3, (uint8_t)motor3);
  ledcWrite(ESC_M4, (uint8_t)motor4);

  if (currentTime % 200 == 0) {
    Serial.print("X:"); Serial.print(angleX, 1);
    Serial.print(" Y:"); Serial.print(angleY, 1);
    Serial.print(" | M:"); Serial.print((int)motor1);
    Serial.print(" "); Serial.print((int)motor2);
    Serial.print(" "); Serial.print((int)motor3);
    Serial.print(" "); Serial.println((int)motor4);
  }

  delay(10);
}

void armESC() {
  for (int i = 0; i < 20; i++) {
    ledcWrite(ESC_M1, 125);
    ledcWrite(ESC_M2, 125);
    ledcWrite(ESC_M3, 125);
    ledcWrite(ESC_M4, 125);
    delay(50);
  }
  Serial.println("ESC armed");
}

float computePID(float current, float target, float kp, float ki, float kd,
                 float &error, float &integral, float &prevError, float dt) {
  error = target - current;
  integral += error * dt;
  integral = constrain(integral, -50, 50);

  float derivative = (error - prevError) / dt;
  prevError = error;

  return kp * error + ki * integral + kd * derivative;
}
