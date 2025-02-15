const randomInt = (min: number, max: number) => {
    return Math.floor(Math.random() * (max - min + 1)) + min;
};

export const randomColor = () => {
    var h = randomInt(0, 360);
    var s = randomInt(20, 98);
    var l = randomInt(20, 90);
    return `hsl(${h},${s}%,${l}%)`;
}
