const randomInt = (min: number, max: number) => {
    return Math.floor(Math.random() * (max - min + 1)) + min;
};

export const NODECOLOR = new Map<number, string>()

export const randomColor = () => {
    var h = randomInt(0, 360);
    var s = randomInt(30, 98);
    var l = randomInt(30, 90);
    return `hsl(${h},${s}%,${l}%)`;
}
