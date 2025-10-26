const loaderPromise = import('./loader.js');

function callLoader(method, ...args) {
  return loaderPromise.then((mod) => {
    if (typeof mod[method] !== 'function') {
      throw new Error(`Loader method not available: ${method}`);
    }
    return mod[method](...args);
  });
}

module.exports.createDWScript = (options = {}) => callLoader('createDWScript', options);
module.exports.ensureRuntimeReady = (options = {}) => callLoader('ensureRuntimeReady', options);
module.exports.getDWScriptClass = () => callLoader('getDWScriptClass');
module.exports.isRuntimeInitialized = () => callLoader('isRuntimeInitialized');
module.exports.resetRuntimeForTesting = () => callLoader('resetRuntimeForTesting');
module.exports.version = '0.1.0';

module.exports.default = module.exports.createDWScript;
