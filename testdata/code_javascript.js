/**
 * JavaScript Code Test File - Tests syntax highlighting for JavaScript language features
 * This file demonstrates various JavaScript language constructs for testing code rendering
 */

'use strict';

// Constants and variables
const API_URL = 'https://api.example.com';
const MAX_RETRIES = 3;
let connectionCount = 0;
var legacyVar = 'still supported';

// Symbol and BigInt
const UNIQUE_KEY = Symbol('unique');
const PRIVATE = Symbol.for('private');
const bigNumber = 9007199254740991n;

// Classes
class EventEmitter {
  #listeners = new Map(); // Private field
  static #instanceCount = 0; // Private static field

  constructor() {
    EventEmitter.#instanceCount++;
  }

  static get instanceCount() {
    return EventEmitter.#instanceCount;
  }

  on(event, callback) {
    if (!this.#listeners.has(event)) {
      this.#listeners.set(event, []);
    }
    this.#listeners.get(event).push(callback);
    return this;
  }

  off(event, callback) {
    if (!this.#listeners.has(event)) return this;
    const callbacks = this.#listeners.get(event);
    const index = callbacks.indexOf(callback);
    if (index > -1) callbacks.splice(index, 1);
    return this;
  }

  emit(event, ...args) {
    if (!this.#listeners.has(event)) return false;
    this.#listeners.get(event).forEach(cb => cb(...args));
    return true;
  }

  once(event, callback) {
    const wrapped = (...args) => {
      this.off(event, wrapped);
      callback(...args);
    };
    return this.on(event, wrapped);
  }
}

// Inheritance
class Logger extends EventEmitter {
  #prefix;

  constructor(prefix = '') {
    super();
    this.#prefix = prefix;
  }

  log(level, message, ...args) {
    const timestamp = new Date().toISOString();
    const formatted = `[${timestamp}] [${level.toUpperCase()}] ${this.#prefix}${message}`;
    console[level](formatted, ...args);
    this.emit('log', { level, message, args, timestamp });
  }

  info(message, ...args) {
    this.log('info', message, ...args);
  }

  warn(message, ...args) {
    this.log('warn', message, ...args);
  }

  error(message, ...args) {
    this.log('error', message, ...args);
  }
}

// Factory function with closure
function createCounter(initialValue = 0) {
  let count = initialValue;

  return {
    get value() {
      return count;
    },
    increment() {
      return ++count;
    },
    decrement() {
      return --count;
    },
    reset() {
      count = initialValue;
      return count;
    },
  };
}

// Higher-order functions
const compose = (...fns) => (x) => fns.reduceRight((acc, fn) => fn(acc), x);
const pipe = (...fns) => (x) => fns.reduce((acc, fn) => fn(acc), x);
const curry = (fn) => {
  const arity = fn.length;
  return function curried(...args) {
    if (args.length >= arity) {
      return fn(...args);
    }
    return (...moreArgs) => curried(...args, ...moreArgs);
  };
};

// Memoization
function memoize(fn) {
  const cache = new Map();
  return function (...args) {
    const key = JSON.stringify(args);
    if (cache.has(key)) {
      return cache.get(key);
    }
    const result = fn.apply(this, args);
    cache.set(key, result);
    return result;
  };
}

// Async/await with error handling
async function fetchWithRetry(url, options = {}, retries = MAX_RETRIES) {
  for (let attempt = 0; attempt < retries; attempt++) {
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      const isLastAttempt = attempt === retries - 1;

      if (isLastAttempt) {
        throw error;
      }

      const delay = Math.pow(2, attempt) * 1000;
      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  }
}

// Promise utilities
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const timeout = (promise, ms) =>
  Promise.race([
    promise,
    new Promise((_, reject) =>
      setTimeout(() => reject(new Error('Timeout')), ms)
    ),
  ]);

const retry = async (fn, retries = 3, delayMs = 1000) => {
  try {
    return await fn();
  } catch (error) {
    if (retries <= 0) throw error;
    await delay(delayMs);
    return retry(fn, retries - 1, delayMs * 2);
  }
};

// Generator functions
function* range(start, end, step = 1) {
  for (let i = start; i < end; i += step) {
    yield i;
  }
}

function* fibonacci() {
  let [a, b] = [0, 1];
  while (true) {
    yield a;
    [a, b] = [b, a + b];
  }
}

// Async generator
async function* asyncRange(start, end, delayMs = 100) {
  for (let i = start; i < end; i++) {
    await delay(delayMs);
    yield i;
  }
}

// Proxy and Reflect
function createObservable(target, onChange) {
  return new Proxy(target, {
    get(obj, prop, receiver) {
      const value = Reflect.get(obj, prop, receiver);
      return typeof value === 'object' && value !== null
        ? createObservable(value, onChange)
        : value;
    },
    set(obj, prop, value, receiver) {
      const oldValue = Reflect.get(obj, prop, receiver);
      const result = Reflect.set(obj, prop, value, receiver);
      if (result && oldValue !== value) {
        onChange(prop, value, oldValue);
      }
      return result;
    },
    deleteProperty(obj, prop) {
      const result = Reflect.deleteProperty(obj, prop);
      if (result) {
        onChange(prop, undefined, obj[prop]);
      }
      return result;
    },
  });
}

// WeakMap and WeakSet
const privateData = new WeakMap();
const processedObjects = new WeakSet();

class SecureContainer {
  constructor(secret) {
    privateData.set(this, { secret });
  }

  getSecret() {
    return privateData.get(this)?.secret;
  }

  process() {
    if (processedObjects.has(this)) {
      throw new Error('Already processed');
    }
    processedObjects.add(this);
    return true;
  }
}

// Destructuring patterns
function processUser({ name, email, settings: { theme = 'light', notifications = true } = {} }) {
  return { name, email, theme, notifications };
}

// Rest and spread
function merge(...objects) {
  return objects.reduce((acc, obj) => ({ ...acc, ...obj }), {});
}

const [first, second, ...rest] = [1, 2, 3, 4, 5];
const { a, b, ...remaining } = { a: 1, b: 2, c: 3, d: 4 };

// Template literals
const createHTML = (title, content) => `
  <!DOCTYPE html>
  <html>
    <head>
      <title>${title}</title>
    </head>
    <body>
      <h1>${title}</h1>
      <div>${content.replace(/</g, '&lt;')}</div>
    </body>
  </html>
`;

// Tagged template literals
function sql(strings, ...values) {
  return {
    text: strings.reduce((acc, str, i) => {
      return acc + str + (i < values.length ? `$${i + 1}` : '');
    }, ''),
    values,
  };
}

const userId = 123;
const query = sql`SELECT * FROM users WHERE id = ${userId}`;

// Array methods
const numbers = [1, 2, 3, 4, 5];
const doubled = numbers.map((n) => n * 2);
const evens = numbers.filter((n) => n % 2 === 0);
const sum = numbers.reduce((acc, n) => acc + n, 0);
const hasEven = numbers.some((n) => n % 2 === 0);
const allPositive = numbers.every((n) => n > 0);
const found = numbers.find((n) => n > 3);
const foundIndex = numbers.findIndex((n) => n > 3);
const flat = [[1, 2], [3, 4], [5]].flat();
const flatMapped = numbers.flatMap((n) => [n, n * 2]);

// Object methods
const obj = { a: 1, b: 2, c: 3 };
const entries = Object.entries(obj);
const keys = Object.keys(obj);
const values = Object.values(obj);
const frozen = Object.freeze({ x: 1 });
const sealed = Object.seal({ y: 2 });

// Optional chaining and nullish coalescing
const user = { name: 'John', address: { city: 'NYC' } };
const city = user?.address?.city ?? 'Unknown';
const country = user?.address?.country ?? 'Unknown';
const callback = user?.callback?.();

// Logical assignment operators
let config = {};
config.timeout ??= 30;
config.retries ||= 3;
config.enabled &&= true;

// Regular expressions
const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
const urlRegex = /^https?:\/\/[\w.-]+(?:\.[\w.-]+)*[\w\-._~:/?#[\]@!$&'()*+,;=]*$/;
const namedGroups = /(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})/;
const match = '2024-01-15'.match(namedGroups);
const { groups: { year, month, day } } = match ?? { groups: {} };

// Error handling
class CustomError extends Error {
  constructor(message, code) {
    super(message);
    this.name = 'CustomError';
    this.code = code;
    Error.captureStackTrace?.(this, CustomError);
  }
}

try {
  throw new CustomError('Something went wrong', 'ERR_CUSTOM');
} catch (error) {
  if (error instanceof CustomError) {
    console.error(`[${error.code}] ${error.message}`);
  } else {
    throw error;
  }
} finally {
  console.log('Cleanup complete');
}

// Main execution
(async function main() {
  const logger = new Logger('[App] ');

  logger.on('log', ({ level, message }) => {
    console.log(`Logged: ${level} - ${message}`);
  });

  logger.info('Application starting...');

  const counter = createCounter(10);
  console.log('Counter:', counter.value, counter.increment(), counter.decrement());

  // Using generators
  for (const n of range(0, 5)) {
    console.log('Range:', n);
  }

  const fib = fibonacci();
  console.log('Fibonacci:', [...Array(10)].map(() => fib.next().value));

  // Async iteration
  for await (const n of asyncRange(0, 3, 10)) {
    console.log('Async:', n);
  }

  logger.info('Application complete');
})();

// Module exports (ES modules syntax - for reference)
export {
  EventEmitter,
  Logger,
  createCounter,
  compose,
  pipe,
  curry,
  memoize,
  fetchWithRetry,
  delay,
  timeout,
  retry,
  createObservable,
  SecureContainer,
};
