import { useEffect, useMemo, useState } from 'react';
import { createDWScript } from '@cwbudde/dwscript';

export default function DWScriptReactExample() {
    const [output, setOutput] = useState('');
    const [code, setCode] = useState("PrintLn('React + DWScript');");
    const [ready, setReady] = useState(false);
    const dwsRef = useMemo(() => ({ current: null }), []);

    useEffect(() => {
        let cancelled = false;

        (async () => {
            const instance = await createDWScript({
                initOptions: {
                    onOutput: (text) => !cancelled && setOutput((prev) => prev + text),
                },
            });

            if (!cancelled) {
                dwsRef.current = instance;
                setReady(true);
            }
        })();

        return () => {
            cancelled = true;
            dwsRef.current?.dispose();
        };
    }, [dwsRef]);

    const runCode = () => {
        if (!ready || !dwsRef.current) {
            return;
        }
        setOutput('');
        const result = dwsRef.current.eval(code);
        if (!result.success && result.error) {
            setOutput(result.error.message);
        }
    };

    return (
        <section>
            <textarea
                rows={8}
                value={code}
                onChange={(event) => setCode(event.target.value)}
                disabled={!ready}
            />
            <button onClick={runCode} disabled={!ready}>
                {ready ? 'Run DWScript' : 'Loading WASM...'}
            </button>
            <pre>{output}</pre>
        </section>
    );
}
