// Type definitions v1

export interface User {
  id: number;
  name: string;
  email: string;
}

export type UserRole = 'admin' | 'user';

export interface Config {
  apiUrl: string;
  timeout: number;
}

export function getUser(id: number): User | null {
  return null;
}

export class UserService {
  private users: User[] = [];

  add(user: User): void {
    this.users.push(user);
  }

  find(id: number): User | undefined {
    return this.users.find(u => u.id === id);
  }
}
