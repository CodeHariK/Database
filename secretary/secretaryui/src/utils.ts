const randomInt = (min: number, max: number) => {
    return Math.floor(Math.random() * (max - min + 1)) + min;
};

export const randomColor = (dark: boolean) => {
    var h = randomInt(0, 360);
    var s = randomInt(20, dark ? 40 : 98);
    var l = randomInt(dark ? 20 : 70, dark ? 40 : 90);
    return `hsl(${h},${s}%,${l}%)`;
}
