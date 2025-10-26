import createDWScript from '@cwbudde/dwscript';

async function main() {
    const dws = await createDWScript({
        initOptions: {
            onOutput: (text) => process.stdout.write(text),
            onError: (error) => console.error('DWScript error:', error.message),
        },
    });

    const program = dws.compile(`
        var i: Integer;
        for i := 1 to 3 do
            PrintLn('Hello from Node #' + IntToStr(i));
    `);

    const result = dws.run(program);
    if (!result.success) {
        console.error(result.error?.message ?? 'Unknown error');
        process.exitCode = 1;
        return;
    }

    console.log('\nExecution time:', result.executionTime, 'ms');

    dws.dispose();
}

main().catch((error) => {
    console.error('Failed to run DWScript example:', error);
    process.exitCode = 1;
});
