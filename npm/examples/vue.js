import { onBeforeUnmount, onMounted, ref } from 'vue';
import { createDWScript } from '@cwbudde/dwscript';

export default {
    name: 'DwscriptVueExample',
    setup() {
        const code = ref("PrintLn('Vue + DWScript');");
        const output = ref('');
        const isReady = ref(false);
        const status = ref('Booting WebAssembly...');
        let instance = null;

        onMounted(async () => {
            instance = await createDWScript({
                initOptions: {
                    onOutput: (text) => (output.value += text),
                    onError: (error) => (status.value = `Runtime error: ${error.message}`),
                },
            });
            status.value = 'Ready';
            isReady.value = true;
        });

        onBeforeUnmount(() => {
            instance?.dispose();
            instance = null;
        });

        const runCode = () => {
            if (!instance) {
                return;
            }
            output.value = '';
            const result = instance.eval(code.value);
            if (!result.success && result.error) {
                status.value = result.error.message;
            } else {
                status.value = `Completed in ${result.executionTime} ms`;
            }
        };

        return { code, output, status, isReady, runCode };
    },
    template: `
        <section class="dwscript-vue-example">
            <p>{{ status }}</p>
            <textarea v-model="code" :disabled="!isReady" rows="8"></textarea>
            <button @click="runCode" :disabled="!isReady">Run DWScript</button>
            <pre>{{ output }}</pre>
        </section>
    `,
};
