import { TrafficSystem } from './traffic.js';

const canvas = document.getElementById('trafficCanvas');
const ctx = canvas.getContext('2d');
const startStopButton = document.getElementById('startStop');
const autoModeToggle = document.getElementById('autoMode');
const lightButtons = document.querySelectorAll('.light-btn');
const densityElement = document.getElementById('density');

function resizeCanvas() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
}

resizeCanvas();
window.addEventListener('resize', resizeCanvas);

const trafficSystem = new TrafficSystem(canvas);
trafficSystem.manualLightChange = true;
let isSimulationRunning = false;
let lastDensityUpdate = 0;

function updateDensity() {
    const totalVehicles = trafficSystem.getTotalVehicles();
    let density;
    if (totalVehicles < 5) density = 'Low';
    else if (totalVehicles < 10) density = 'Medium';
    else density = 'High';

    densityElement.textContent = density;
    densityElement.className = `density-${density.toLowerCase()}`;
}

isSimulationRunning = true;
animate();
startStopButton.addEventListener('click', () => {
    isSimulationRunning = !isSimulationRunning;
    startStopButton.textContent = isSimulationRunning ? 'Stop Simulation' : 'Start Simulation';

    if (isSimulationRunning) {
        animate();
    }
});

document.querySelectorAll('.spawn-btn').forEach(button => {
    button.addEventListener('click', () => {
        const direction = button.dataset.direction;
        trafficSystem.manualSpawn(direction);
    });
});

autoModeToggle.addEventListener('change', () => {
    trafficSystem.manualLightChange = !autoModeToggle.checked;
    lightButtons.forEach(btn => {
        btn.disabled = autoModeToggle.checked;
    });
});

lightButtons.forEach(button => {
    button.addEventListener('click', () => {
        if (!trafficSystem.manualLightChange) return;

        const l = trafficSystem.trafficLights.find(l => l.state === 'green');
        if (l && l.direction !== button.dataset.direction && button.dataset.color === 'green') {
            $.confirm({
                title: 'Change Traffic Light',
                content: `Multiple lights are currently green. Do you want to switch the current green light to red and make ${button.dataset.direction} green?`,
                theme: 'supervan',
                useBootstrap: false,
                closeIcon: true,
                columnClass: 'col-md-4 col-md-offset-4',
                type: 'red',
                buttons: {
                    confirm: {
                        text: 'Yes',
                        action: () => {
                            const light = trafficSystem.trafficLights.find(l => l.direction === button.dataset.direction);
                            if (light) {
                                light.manualChange(button.dataset.color);
                            }
                        },
                        btnClass: 'btn-red'
                    },
                    switch: {
                        text: 'Switch',
                        action: () => {
                            // make current green light -> red
                            l.manualChange('red');
                            // make the new light green
                            const light = trafficSystem.trafficLights.find(l => l.direction === button.dataset.direction);
                            if (light) {
                                light.manualChange(button.dataset.color);
                            }
                        },
                        btnClass: 'btn-green'
                    },
                    cancel: {
                        text: 'No',
                        action: () => {
                            // close the modal
                        },
                        btnClass: 'btn-blue'
                    }
                }
            });
        } else {
            const direction = button.dataset.direction;
            const color = button.dataset.color;
            const light = trafficSystem.trafficLights.find(l => l.direction === direction);
            if (light) {
                light.manualChange(color);
            }
        }
    });
});

lightButtons.forEach(btn => {
    btn.disabled = autoModeToggle.checked;
});

function animate(timestamp) {
    if (!isSimulationRunning) return;

    ctx.clearRect(0, 0, canvas.width, canvas.height);
    trafficSystem.update();
    trafficSystem.draw(ctx);

    document.getElementById('nsCount').textContent = trafficSystem.getNorthSouthCount();
    document.getElementById('ewCount').textContent = trafficSystem.getEastWestCount();

    if (timestamp - lastDensityUpdate > 1000) {
        updateDensity();
        lastDensityUpdate = timestamp;
    }

    requestAnimationFrame(animate);
}

let socket_url = 'ws://localhost/ws';
if (window.location.protocol === 'https:') {
    socket_url = `wss://${window.location.host}/ws`;
}

const socket = new WebSocket(socket_url);
socket.onmessage = (event) => {
    if (!trafficSystem.manualLightChange) return;
    let data;
    try {
        data = JSON.parse(event.data);
    } catch (error) {
        console.error('Invalid JSON data received', event.data);
        return;
    }
    if (data.type === 'vehicleUpdate') {
        trafficSystem.manualSpawn(data.direction);
    } else if (data.type === 'lightChange') {
        trafficSystem.trafficLights.forEach(l => {
            if (l.direction !== data.direction) trafficSystem.ManualChangeLight(l.direction, 'red');
            else l.setDuration(data.timeEffective * 1000 || 10000);
        });
        trafficSystem.ManualChangeLight(data.direction, data.color);
        socket.send(JSON.stringify({ phases: trafficSystem.getVehiclesInAllPhases() }));
    }
};
