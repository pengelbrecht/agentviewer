// TypeScript Code Test File - Tests syntax highlighting for TypeScript language features
// This file demonstrates various TypeScript language constructs for testing code rendering

import { EventEmitter } from 'events';

// Type aliases and utility types
type ID = string | number;
type Nullable<T> = T | null | undefined;
type ReadOnly<T> = { readonly [P in keyof T]: T[P] };
type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

// Enums
enum Status {
  Pending = 'PENDING',
  Running = 'RUNNING',
  Complete = 'COMPLETE',
  Failed = 'FAILED',
}

const enum Direction {
  Up = 1,
  Down,
  Left,
  Right,
}

// Interfaces
interface User {
  id: ID;
  name: string;
  email: string;
  createdAt: Date;
  metadata?: Record<string, unknown>;
}

interface ApiResponse<T> {
  data: T;
  status: number;
  message?: string;
  pagination?: {
    page: number;
    pageSize: number;
    total: number;
  };
}

// Interface extending
interface Admin extends User {
  permissions: string[];
  lastLogin: Date;
}

// Generic constraints
interface Identifiable {
  id: ID;
}

interface Repository<T extends Identifiable> {
  find(id: ID): Promise<T | null>;
  findAll(filter?: Partial<T>): Promise<T[]>;
  create(data: Omit<T, 'id'>): Promise<T>;
  update(id: ID, data: Partial<T>): Promise<T>;
  delete(id: ID): Promise<boolean>;
}

// Classes with decorators (conceptual - decorators need experimental flag)
abstract class Entity {
  abstract id: ID;

  constructor(public createdAt: Date = new Date()) {}

  abstract validate(): boolean;
}

class UserEntity extends Entity implements User {
  id: ID;
  name: string;
  email: string;

  constructor(data: Omit<User, 'createdAt'>) {
    super();
    this.id = data.id;
    this.name = data.name;
    this.email = data.email;
  }

  validate(): boolean {
    return (
      typeof this.name === 'string' &&
      this.name.length > 0 &&
      this.email.includes('@')
    );
  }

  // Getter and setter
  get displayName(): string {
    return `${this.name} <${this.email}>`;
  }

  set displayName(value: string) {
    const match = value.match(/^(.+) <(.+)>$/);
    if (match) {
      this.name = match[1];
      this.email = match[2];
    }
  }

  // Static method
  static fromJSON(json: string): UserEntity {
    const data = JSON.parse(json) as User;
    return new UserEntity(data);
  }

  // Private method
  #hashPassword(password: string): string {
    // Private class field method
    return password.split('').reverse().join('');
  }
}

// Generic class
class AsyncQueue<T> {
  private items: T[] = [];
  private waiters: ((item: T) => void)[] = [];

  async enqueue(item: T): Promise<void> {
    if (this.waiters.length > 0) {
      const waiter = this.waiters.shift()!;
      waiter(item);
    } else {
      this.items.push(item);
    }
  }

  async dequeue(): Promise<T> {
    if (this.items.length > 0) {
      return this.items.shift()!;
    }

    return new Promise<T>((resolve) => {
      this.waiters.push(resolve);
    });
  }

  get size(): number {
    return this.items.length;
  }
}

// Function overloads
function process(input: string): string;
function process(input: number): number;
function process(input: string | number): string | number {
  if (typeof input === 'string') {
    return input.toUpperCase();
  }
  return input * 2;
}

// Arrow functions and type guards
const isUser = (obj: unknown): obj is User => {
  return (
    typeof obj === 'object' &&
    obj !== null &&
    'id' in obj &&
    'name' in obj &&
    'email' in obj
  );
};

// Async/await patterns
async function fetchUser(id: ID): Promise<User | null> {
  try {
    const response = await fetch(`/api/users/${id}`);

    if (!response.ok) {
      throw new Error(`HTTP error: ${response.status}`);
    }

    const data = await response.json();

    if (!isUser(data)) {
      throw new Error('Invalid user data');
    }

    return data;
  } catch (error) {
    if (error instanceof Error) {
      console.error('Failed to fetch user:', error.message);
    }
    return null;
  }
}

// Promise combinators
async function fetchAllUsers(ids: ID[]): Promise<User[]> {
  const results = await Promise.allSettled(
    ids.map(id => fetchUser(id))
  );

  return results
    .filter((r): r is PromiseFulfilledResult<User> =>
      r.status === 'fulfilled' && r.value !== null
    )
    .map(r => r.value);
}

// Mapped types
type EventMap = {
  connect: void;
  disconnect: { reason: string };
  message: { content: string; timestamp: Date };
  error: Error;
};

type EventHandler<T> = T extends void ? () => void : (data: T) => void;

// Conditional types
type UnwrapPromise<T> = T extends Promise<infer U> ? U : T;
type ArrayElement<T> = T extends (infer E)[] ? E : never;

// Template literal types
type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
type Endpoint = `/${string}`;
type RouteKey = `${HTTPMethod} ${Endpoint}`;

// Discriminated union
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

function handleResult<T>(result: Result<T>): T | null {
  if (result.success) {
    return result.data;
  }
  console.error(result.error);
  return null;
}

// Infer keyword usage
type ReturnTypeOf<T> = T extends (...args: any[]) => infer R ? R : never;
type Parameters<T> = T extends (...args: infer P) => any ? P : never;

// keyof and typeof
const CONFIG = {
  apiUrl: 'https://api.example.com',
  timeout: 5000,
  retries: 3,
} as const;

type ConfigKey = keyof typeof CONFIG;
type ConfigValue = typeof CONFIG[ConfigKey];

// Namespace
namespace Validation {
  export interface Rule {
    validate(value: unknown): boolean;
    message: string;
  }

  export const required: Rule = {
    validate: (value) => value !== null && value !== undefined && value !== '',
    message: 'This field is required',
  };

  export function createRule(
    validate: (value: unknown) => boolean,
    message: string
  ): Rule {
    return { validate, message };
  }
}

// Module augmentation
declare module 'events' {
  interface EventEmitter {
    emitAsync(event: string, ...args: unknown[]): Promise<boolean>;
  }
}

// Main execution
async function main(): Promise<void> {
  // Object destructuring with types
  const { apiUrl, timeout }: typeof CONFIG = CONFIG;

  // Array destructuring
  const [first, second, ...rest] = [1, 2, 3, 4, 5];

  // Nullish coalescing and optional chaining
  const user = await fetchUser('123');
  const userName = user?.name ?? 'Anonymous';
  const emailDomain = user?.email?.split('@')[1];

  // Template literals
  const greeting = `Hello, ${userName}!`;

  // Spread operator
  const extended = { ...user, role: 'admin' };

  // Tagged template literal
  function sql(strings: TemplateStringsArray, ...values: unknown[]): string {
    return strings.reduce((result, str, i) => {
      const value = values[i] !== undefined ? String(values[i]) : '';
      return result + str + value;
    }, '');
  }

  const query = sql`SELECT * FROM users WHERE id = ${123}`;

  console.log(greeting, query, apiUrl, timeout, first, second, rest, emailDomain, extended);
}

main().catch(console.error);

// Export
export { User, Admin, Status, UserEntity, AsyncQueue, fetchUser, Validation };
export type { ID, ApiResponse, Repository, Result };
