class Vehicle {
    constructor(x, y, direction, lane) {
        this.x = x;
        this.y = y;
        this.direction = direction;
        this.lane = lane;
        this.speed = 3;
        this.width = 30;
        this.height = 50;
        this.color = ['#FF0000', '#00FF00', '#0000FF', '#FFFF00'][Math.floor(Math.random() * 4)];
        this.turnDirection = 'straight' //Math.random() < 0.3 ? (Math.random() < 0.5 ? 'left' : 'right') : 'straight';
        this.turning = false;
        this.turnProgress = 0;
        this.inIntersection = false;
        this.waitingAtLight = false;
        this.turnPoint = null;
        this.turnRadius = 60;
        this.turnAngle = 0;
        this.acceleration = 0.1;
        this.maxSpeed = 4;
        this.brakeForce = 0.2;
    }

    IsVehicleInIntersection(vehicle) {
        const intersectionSize = 100;
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;
        return this.vehicles.some((vehicle) => {
            const dx = vehicle.x - centerX;
            const dy = vehicle.y - centerY;
            return Math.abs(dx) < intersectionSize && Math.abs(dy) < intersectionSize;
        });
    }

    update(trafficLight, vehicles) {
        if (this.waitingAtLight) {
            if (trafficLight.state === 'green' && !this.isVehicleAhead(vehicles, 40)) {
                this.waitingAtLight = false;
            }
            return;
        }

        const aheadDistance = this.isVehicleAhead(vehicles, 100);
        if (aheadDistance) {
            this.speed = Math.max(this.speed - this.brakeForce, 0);
        } else {
            this.speed = Math.min(this.speed + this.acceleration, this.maxSpeed);
        }

        if (this.turning) {
            this.executeTurn();
        } else {
            const shouldStop = this.checkTrafficLight(trafficLight);
            if (shouldStop) {
                this.waitingAtLight = true;
                return;
            }

            if (!this.isVehicleAhead(vehicles, this.speed + 30)) {
                this.moveForward();
            }

            if (!this.inIntersection) {
                this.checkIntersection(trafficLight);
            }
        }


    }

    isVehicleAhead(vehicles, distance) {
        const lookAheadDistance = distance;
        const sideDistance = 15;

        const myOffset = this.getLaneOffset();
        const myX = this.x + myOffset.x;
        const myY = this.y + myOffset.y;

        return vehicles.some((other) => {
            if (other === this) return false;

            const otherOffset = other.getLaneOffset();
            const otherX = other.x + otherOffset.x;
            const otherY = other.y + otherOffset.y;

            const dx = otherX - myX;
            const dy = otherY - myY;

            if (other.direction !== this.direction || other.lane !== this.lane) return false;

            switch (this.direction) {
                case 'north':
                    return Math.abs(dx) < sideDistance && dy < 0 && Math.abs(dy) < lookAheadDistance;
                case 'south':
                    return Math.abs(dx) < sideDistance && dy > 0 && Math.abs(dy) < lookAheadDistance;
                case 'east':
                    return Math.abs(dy) < sideDistance && dx > 0 && Math.abs(dx) < lookAheadDistance;
                case 'west':
                    return Math.abs(dy) < sideDistance && dx < 0 && Math.abs(dx) < lookAheadDistance;
            }
        });
    }

    getLaneOffset() {
        const laneWidth = 40;
        switch (this.direction) {
            case 'north':
                return { x: this.lane === 0 ? -laneWidth : laneWidth, y: 0 };
            case 'south':
                return { x: this.lane === 1 ? laneWidth : -laneWidth, y: 0 };
            case 'east':
                return { x: 0, y: this.lane === 0 ? -laneWidth : laneWidth };
            case 'west':
                return { x: 0, y: this.lane === 1 ? laneWidth : -laneWidth };
            default:
                return { x: 0, y: 0 };
        }
    }

    moveForward() {
        const speedFactor = this.turning ? 0.7 : 1;
        switch (this.direction) {
            case 'north':
                this.y -= this.speed * speedFactor;
                break;
            case 'south':
                this.y += this.speed * speedFactor;
                break;
            case 'east':
                this.x += this.speed * speedFactor;
                break;
            case 'west':
                this.x -= this.speed * speedFactor;
                break;
        }
    }

    checkIntersection(trafficLight) {
        if (this.turnDirection === 'straight') {
            this.inIntersection = true;
            return;
        }

        const intersectionSize = 100;
        const dx = this.x - trafficLight.centerX;
        const dy = this.y - trafficLight.centerY;

        if (Math.abs(dx) < intersectionSize && Math.abs(dy) < intersectionSize) {
            this.turning = true;
            this.turnPoint = {
                x: trafficLight.centerX + (this.turnDirection === 'left' ? -40 : 40),
                y: trafficLight.centerY + (this.turnDirection === 'left' ? -40 : 40),
            };
            this.inIntersection = true;
            this.initializeTurnParameters();
        }

        console.log(this.inIntersection);
    }

    initializeTurnParameters() {
        this.turning = true;
        this.turnProgress = 0;
        this.speed *= 0.7;

        const leftTurn = -Math.PI / 2;
        const rightTurn = Math.PI / 2;

        const directions = {
            'north': 0,
            'east': Math.PI / 2,
            'south': Math.PI,
            'west': -Math.PI / 2
        };

        this.startAngle = directions[this.direction];
        this.targetAngle = this.startAngle + (this.turnDirection === 'left' ? leftTurn : rightTurn);

        // Ensure correct pivot point for turning based on lanes
        let laneOffset = 20; // Assuming lane width is 40, offset by half
        let intersectionOffset = 50; // Half of intersection size

        if (this.lane === 'left') {
            laneOffset *= -1;
        }

        switch (this.direction) {
            case 'north':
                this.turnPoint = {
                    x: this.x + (this.turnDirection === 'left' ? (-intersectionOffset + laneOffset) : (intersectionOffset + laneOffset)),
                    y: this.y - intersectionOffset
                };
                break;
            case 'south':
                this.turnPoint = {
                    x: this.x + (this.turnDirection === 'left' ? (intersectionOffset + laneOffset) : (-intersectionOffset + laneOffset)),
                    y: this.y + intersectionOffset
                };
                break;
            case 'east':
                this.turnPoint = {
                    x: this.x + intersectionOffset,
                    y: this.y + (this.turnDirection === 'left' ? (intersectionOffset + laneOffset) : (-intersectionOffset + laneOffset))
                };
                break;
            case 'west':
                this.turnPoint = {
                    x: this.x - intersectionOffset,
                    y: this.y + (this.turnDirection === 'left' ? (-intersectionOffset + laneOffset) : (intersectionOffset + laneOffset))
                };
                break;
        }

        this.turnRadius = 50; // Half of intersection size
    }

    executeTurn() {
        const turnSpeed = Math.PI / 90;
        this.turnProgress += turnSpeed;

        if (this.turnProgress >= Math.PI / 2) {
            this.turnProgress = Math.PI / 2;
            this.turning = false;
            this.updateDirectionAfterTurn();
            this.speed = Math.min(this.speed * 1.4, this.maxSpeed);
        }

        const angle = this.startAngle + this.turnProgress;
        this.x = this.turnPoint.x + this.turnRadius * Math.cos(angle);
        this.y = this.turnPoint.y + this.turnRadius * Math.sin(angle);
    }

    updateDirectionAfterTurn() {
        const directions = ['north', 'east', 'south', 'west'];
        let currentIndex = directions.indexOf(this.direction);

        if (this.turnDirection === 'left') {
            currentIndex = (currentIndex - 1 + 4) % 4;
        } else if (this.turnDirection === 'right') {
            currentIndex = (currentIndex + 1) % 4;
        }

        this.direction = directions[currentIndex];
        this.lane = 'left'; // Maintain proper lane after turn
        this.turning = false;
        this.turnProgress = 0;
    }

    checkTrafficLight(trafficLight) {
        if (trafficLight.state === 'red' || (trafficLight.state === 'yellow' && !this.inIntersection)) {
            const stopBuffer = 100;
            const nearStopLineDistance = 100;
            const centerX = trafficLight.centerX;
            const centerY = trafficLight.centerY;

            let isNearStopLine = false;
            switch (this.direction) {
                case 'north':
                    isNearStopLine = this.y > centerY + stopBuffer && this.y < centerY + stopBuffer + nearStopLineDistance;
                    break;
                case 'south':
                    isNearStopLine = this.y < centerY - stopBuffer && this.y > centerY - stopBuffer - nearStopLineDistance;
                    break;
                case 'east':
                    isNearStopLine = this.x < centerX - stopBuffer && this.x > centerX - stopBuffer - nearStopLineDistance;
                    break;
                case 'west':
                    isNearStopLine = this.x > centerX + stopBuffer && this.x < centerX + stopBuffer + nearStopLineDistance;
                    break;
            }

            if (isNearStopLine) {
                switch (this.direction) {
                    case 'north':
                        return this.y > centerY + stopBuffer;
                    case 'south':
                        return this.y < centerY - stopBuffer;
                    case 'east':
                        return this.x < centerX - stopBuffer;
                    case 'west':
                        return this.x > centerX + stopBuffer;
                }
            }
        }
        return false;
    }

    draw(ctx) {
        const offset = this.getLaneOffset();

        ctx.save();
        ctx.translate(this.x + offset.x, this.y + offset.y);

        let angle = 0;
        switch (this.direction) {
            case 'north': angle = 0; break;
            case 'east': angle = Math.PI / 2; break;
            case 'south': angle = Math.PI; break;
            case 'west': angle = -Math.PI / 2; break;
        }

        if (this.turning) {
            angle += this.turnProgress;
        }

        ctx.shadowColor = 'rgba(0, 0, 0, 0.3)';
        ctx.shadowBlur = 10;
        ctx.shadowOffsetX = 5;
        ctx.shadowOffsetY = 5;

        ctx.rotate(angle);

        ctx.fillStyle = this.color;
        ctx.fillRect(-this.width / 2, -this.height / 2, this.width, this.height);

        ctx.fillStyle = '#87CEEB';
        ctx.fillRect(-this.width / 2 + 5, -this.height / 2 + 5, this.width - 10, this.height / 3);

        if (this.turnDirection !== 'straight') {
            ctx.fillStyle = '#FFFF00';
            const indicatorSize = 5;
            if (this.turnDirection === 'left') {
                ctx.fillRect(-this.width / 2, -this.height / 4, indicatorSize, indicatorSize);
            } else {
                ctx.fillRect(this.width / 2 - indicatorSize, -this.height / 4, indicatorSize, indicatorSize);
            }
        }

        ctx.fillStyle = '#FFFFFF';
        ctx.fillRect(-this.width / 2 + 2, -this.height / 2, 8, 5);
        ctx.fillRect(this.width / 2 - 10, -this.height / 2, 8, 5);

        ctx.restore();
    }
}

class TrafficLight {
    constructor(x, y, direction, centerX, centerY) {
        this.position = { x, y };
        this.direction = direction;
        this.state = 'red';
        this.centerX = centerX;
        this.centerY = centerY;
        this.yellowTimer = 0;
        this.isChanging = false;
        this.greenDuration = 100;
        this.durDecFunc;
        this.phaseTimer = 0;
        this.yellowDuration = 40;
    }

    manualChange(newState) {
        if (this.state !== newState) {
            if (newState === 'green' && this.state === 'red') {
                this.state = newState;
            } else if (newState === 'red' && this.state === 'green') {
                this.state = 'yellow';
                this.isChanging = true;
                setTimeout(() => {
                    this.state = newState;
                    this.isChanging = false;
                }, this.yellowDuration * 10);
            } else if (newState === 'yellow' && this.state === 'green') {
                this.state = newState;
                this.isChanging = true;
                setTimeout(() => {
                    this.state = 'red';
                    this.isChanging = false;
                }, this.yellowDuration * 10);
            } else if (newState === 'yellow' && this.state === 'red') {
                this.state = newState;
                this.isChanging = true;
                setTimeout(() => {
                    this.state = 'green';
                    this.isChanging = false;
                }, this.yellowDuration * 10);
            }
        }
    }

    setDuration(greenDuration) {
        this.greenDuration = greenDuration
        if (this.durDecFunc == null) {
            this.durDecFunc = setInterval(() => {
                this.greenDuration -= 1000;
            }, 1000);
        }
    }

    update(phase) {
        if (this.isChanging) {
            this.yellowTimer++;
            if (this.yellowTimer >= this.yellowDuration) {
                this.isChanging = false;
                this.yellowTimer = 0;
                this.state = 'red';
            }
            return;
        }

        const shouldBeGreen =
            (phase === 0 && this.direction === 'north') ||
            (phase === 1 && this.direction === 'south') ||
            (phase === 2 && this.direction === 'east') ||
            (phase === 3 && this.direction === 'west');

        if (shouldBeGreen && this.state === 'red') {
            this.state = 'green';
            this.phaseTimer = 0;
        } else if (!shouldBeGreen && this.state === 'green') {
            this.state = 'yellow';
            this.isChanging = true;
        }

        this.phaseTimer++;
    }

    draw(ctx) {
        ctx.save();
        ctx.fillStyle = '#333';
        ctx.fillRect(this.position.x - 15, this.position.y - 15, 30, 30);

        ctx.fillStyle = this.state === 'red' ? '#FF0000' :
            this.state === 'yellow' ? '#FFFF00' : '#00FF00';

        ctx.beginPath();
        ctx.arc(this.position.x, this.position.y, 10, 0, Math.PI * 2);
        ctx.fill();

        ctx.shadowColor = this.state === 'red' ? '#FF0000' :
            this.state === 'yellow' ? '#FFFF00' : '#00FF00';
        ctx.shadowBlur = 20;
        ctx.fill();

        if (this.state !== 'yellow' && this.state !== 'red') {
            const remainingTime = Math.ceil(this.greenDuration / 1000);

            if (remainingTime > 0) {
                ctx.font = '14px Inter';
                ctx.fillStyle = '#FFFFFF';
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';

                const offset = 30;
                let textX = this.position.x;
                let textY = this.position.y;

                switch (this.direction) {
                    case 'north':
                        textY += offset;
                        break;
                    case 'south':
                        textY -= offset;
                        break;
                    case 'east':
                        textX -= offset;
                        textY += 2;
                        break;
                    case 'west':
                        textX += offset;
                        textY += 2;
                        break;
                }

                ctx.fillText('(' + remainingTime + 's)', textX, textY);
            }
        }

        ctx.restore();
    }
}

export class TrafficSystem {
    constructor(canvas) {
        this.canvas = canvas;
        this.vehicles = [];
        this.trafficLights = this.setupTrafficLights();
        this.vehicleSpawnTimer = 0;
        this.spawnInterval = 240;
        this.northSouthCount = 0;
        this.eastWestCount = 0;
        this.phase = 0;
        this.phaseTimer = 0;
        this.phaseDuration = 400;
        this.maxVehiclesPerDirection = 4;
        this.manualLightChange = true;
        this.dynamicPhasing = true;
        this.grassPattern = this.createGrassPattern();

        this.vehiclesInPhases = [0, 0, 0, 0];
    }

    createGrassPattern() {
        const patternCanvas = document.createElement('canvas');
        patternCanvas.width = 22;
        patternCanvas.height = 22;
        const pctx = patternCanvas.getContext('2d');

        pctx.fillStyle = '#1a472a';
        pctx.fillRect(0, 0, 20, 20);

        pctx.fillStyle = '#2d5a3d';
        for (let i = 0; i < 8; i++) {
            const x = Math.random() * 20;
            const y = Math.random() * 20;
            const size = 1 + Math.random() * 2;
            pctx.beginPath();
            pctx.arc(x, y, size, 0, Math.PI * 2);
            pctx.fill();
        }

        return pctx.createPattern(patternCanvas, 'repeat');
    }

    setupTrafficLights() {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;
        const intersectionSize = 100;

        return [
            new TrafficLight(centerX, centerY - intersectionSize, 'north', centerX, centerY),
            new TrafficLight(centerX, centerY + intersectionSize, 'south', centerX, centerY),
            new TrafficLight(centerX - intersectionSize, centerY, 'west', centerX, centerY),
            new TrafficLight(centerX + intersectionSize, centerY, 'east', centerX, centerY),
        ];
    }

    manualSpawn(direction) {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;
        let lane = Math.random() < 0.5 ? 0 : 1;

        let x, y;
        switch (direction) {
            case 'north':
                x = centerX;
                y = this.canvas.height + 50;
                this.northSouthCount++;
                lane = 0;
                break;
            case 'south':
                x = centerX;
                y = -50;
                this.northSouthCount++;
                lane = 1;
                break;
            case 'east':
                x = -50;
                y = centerY;
                this.eastWestCount++;
                lane = 0;
                break;
            case 'west':
                x = this.canvas.width + 50;
                y = centerY;
                this.eastWestCount++;
                lane = 1;
                break;
        }

        this.vehicles.push(new Vehicle(x, y, direction, lane));
    }

    ManualChangeLight(direction, state) {
        if (state === 'green') {
            setTimeout(() => {
                this.trafficLights.find(light => light.direction === direction).manualChange(state);
            }, 400);
        } else {
            this.trafficLights.find(light => light.direction === direction).manualChange(state);
        }
    }


    IntersectionBuzy() {
        const intersectionSize = 100;
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;
        return this.vehicles.some((vehicle) => {
            const dx = vehicle.x - centerX;
            const dy = vehicle.y - centerY;
            return Math.abs(dx) < intersectionSize && Math.abs(dy) < intersectionSize;
        });
    }

    update() {
        this.phaseTimer++;

        if (this.dynamicPhasing) {
            const vehiclesInPhase = this.getVehiclesInCurrentPhase();
            this.vehiclesInPhases[this.phase] = vehiclesInPhase;
            this.phaseDuration = Math.max(200, Math.min(600, vehiclesInPhase * 100));
        }

        if (this.phaseTimer >= this.phaseDuration) {
            this.phase = (this.phase + 1) % 4;
            this.phaseTimer = 0;
        }

        if (!this.manualLightChange) {
            for (this.IntersectionBuzy(); this.phaseTimer >= this.phaseDuration; this.phase = (this.phase + 1) % 4) {
                this.phaseTimer = 0;
            }
            this.trafficLights.forEach((light) => light.update(this.phase));
        }

        this.vehicles.forEach((vehicle) => {
            const relevantLight = this.trafficLights.find(
                (light) => light.direction === vehicle.direction
            );
            vehicle.update(relevantLight, this.vehicles);
        });

        this.vehicles = this.vehicles.filter((vehicle) => {
            return !(
                vehicle.x < -100 ||
                vehicle.x > this.canvas.width + 100 ||
                vehicle.y < -100 ||
                vehicle.y > this.canvas.height + 100
            );
        });
    }

    getVehiclesInCurrentPhase() {
        let count = 0;
        const currentDirection = ['north', 'south', 'east', 'west'][this.phase];
        this.vehicles.forEach(vehicle => {
            if (vehicle.direction === currentDirection) count++;
        });
        return count;
    }

    getVehiclesInAllPhases() {
        let phaseMap = [0, 0, 0, 0];
        this.vehicles.forEach(vehicle => {
            switch (vehicle.direction) {
                case 'north': phaseMap[0]++; break;
                case 'south': phaseMap[1]++; break;
                case 'east': phaseMap[2]++; break;
                case 'west': phaseMap[3]++; break;
            }
        })
        return phaseMap;
    }

    getTotalVehicles() {
        return this.vehicles.length;
    }

    draw(ctx) {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;
        const roadWidth = 200;
        const laneWidth = 40;
        const grassWidth = 60;

        ctx.save();

        ctx.fillStyle = this.grassPattern;

        ctx.fillRect(centerX - roadWidth / 2 - grassWidth, 0, grassWidth, this.canvas.height);
        ctx.fillRect(centerX + roadWidth / 2, 0, grassWidth, this.canvas.height);

        ctx.fillRect(0, centerY - roadWidth / 2 - grassWidth, this.canvas.width, grassWidth);
        ctx.fillRect(0, centerY + roadWidth / 2, this.canvas.width, grassWidth);
        ctx.globalAlpha = 0.1;
        for (let i = 0; i < this.canvas.width; i += 30) {
            for (let j = 0; j < this.canvas.height; j += 30) {
                if (Math.random() > 0.5) {
                    ctx.fillStyle = '#ffffff';
                    ctx.fillRect(i, j, 2, 2);
                }
            }
        }
        ctx.globalAlpha = 1;

        ctx.fillStyle = '#333';
        ctx.fillRect(centerX - roadWidth / 2, 0, roadWidth, this.canvas.height);
        ctx.fillRect(0, centerY - roadWidth / 2, this.canvas.width, roadWidth);

        ctx.strokeStyle = '#ff3333';
        ctx.lineWidth = 2;
        ctx.shadowColor = '#ff3333';
        ctx.shadowBlur = 10;

        ctx.beginPath();
        ctx.moveTo(centerX - roadWidth / 2, 0);
        ctx.lineTo(centerX - roadWidth / 2, this.canvas.height);
        ctx.moveTo(centerX + roadWidth / 2, 0);
        ctx.lineTo(centerX + roadWidth / 2, this.canvas.height);
        ctx.stroke();

        ctx.beginPath();
        ctx.moveTo(0, centerY - roadWidth / 2);
        ctx.lineTo(this.canvas.width, centerY - roadWidth / 2);
        ctx.moveTo(0, centerY + roadWidth / 2);
        ctx.lineTo(this.canvas.width, centerY + roadWidth / 2);
        ctx.stroke();

        ctx.shadowBlur = 0;

        ctx.strokeStyle = '#FFF';
        ctx.setLineDash([20, 20]);

        ctx.beginPath();
        ctx.moveTo(centerX - laneWidth, 0);
        ctx.lineTo(centerX - laneWidth, this.canvas.height);
        ctx.moveTo(centerX + laneWidth, 0);
        ctx.lineTo(centerX + laneWidth, this.canvas.height);
        ctx.stroke();

        ctx.beginPath();
        ctx.moveTo(0, centerY - laneWidth);
        ctx.lineTo(this.canvas.width, centerY - laneWidth);
        ctx.moveTo(0, centerY + laneWidth);
        ctx.lineTo(this.canvas.width, centerY + laneWidth);
        ctx.stroke();

        ctx.setLineDash([]);

        ctx.strokeStyle = '#FFFF00';
        ctx.lineWidth = 3;
        ctx.beginPath();
        ctx.moveTo(centerX - roadWidth / 2, centerY - roadWidth / 2);
        ctx.lineTo(centerX + roadWidth / 2, centerY - roadWidth / 2);
        ctx.moveTo(centerX - roadWidth / 2, centerY + roadWidth / 2);
        ctx.lineTo(centerX + roadWidth / 2, centerY + roadWidth / 2);
        ctx.moveTo(centerX + roadWidth / 2, centerY - roadWidth / 2);
        ctx.lineTo(centerX + roadWidth / 2, centerY + roadWidth / 2);
        ctx.moveTo(centerX - roadWidth / 2, centerY - roadWidth / 2);
        ctx.lineTo(centerX - roadWidth / 2, centerY + roadWidth / 2);
        ctx.stroke();

        ctx.restore();
        this.trafficLights.forEach((light) => light.draw(ctx));
        this.vehicles.forEach((vehicle) => vehicle.draw(ctx));

        ctx.save();
        ctx.font = '16px Inter';
        ctx.fillStyle = '#FFFFFF';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';

        let northCount = 0, southCount = 0, eastCount = 0, westCount = 0;
        this.vehicles.forEach(vehicle => {
            switch (vehicle.direction) {
                case 'north': northCount++; break;
                case 'south': southCount++; break;
                case 'east': eastCount++; break;
                case 'west': westCount++; break;
            }
        });

        const countOffset = 330;

        ctx.fillText(`${northCount}`, centerX, centerY + countOffset);
        ctx.fillText(`${southCount}`, centerX, centerY - countOffset);
        ctx.fillText(`${eastCount} ->`, centerX - countOffset, centerY);
        ctx.fillText(`${westCount} <-`, centerX + countOffset, centerY);
    }

    getNorthSouthCount() {
        return this.northSouthCount;
    }

    getEastWestCount() {
        return this.eastWestCount;
    }
}