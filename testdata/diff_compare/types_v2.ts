// Type definitions v2 (modernized)

export interface BaseEntity {
  id: string;  // Changed from number to string (UUID)
  createdAt: Date;
  updatedAt: Date;
}

export interface User extends BaseEntity {
  name: string;
  email: string;
  role: UserRole;
  isActive: boolean;
  metadata?: Record<string, unknown>;
}

export type UserRole = 'admin' | 'moderator' | 'user' | 'guest';

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

export interface Config {
  apiUrl: string;
  timeout: number;
  retries: number;
  features: {
    darkMode: boolean;
    notifications: boolean;
  };
}

export async function getUser(id: string): Promise<User | null> {
  // Now async with string ID
  return null;
}

export async function getUsers(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<User>> {
  return {
    data: [],
    total: 0,
    page,
    pageSize,
    hasMore: false,
  };
}

export class UserService {
  private users: Map<string, User> = new Map();

  async add(user: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): Promise<User> {
    const now = new Date();
    const newUser: User = {
      ...user,
      id: crypto.randomUUID(),
      createdAt: now,
      updatedAt: now,
    };
    this.users.set(newUser.id, newUser);
    return newUser;
  }

  async find(id: string): Promise<User | undefined> {
    return this.users.get(id);
  }

  async findByEmail(email: string): Promise<User | undefined> {
    return Array.from(this.users.values()).find(u => u.email === email);
  }

  async update(id: string, updates: Partial<User>): Promise<User | null> {
    const user = this.users.get(id);
    if (!user) return null;

    const updated: User = {
      ...user,
      ...updates,
      id: user.id,  // Prevent ID modification
      updatedAt: new Date(),
    };
    this.users.set(id, updated);
    return updated;
  }

  async delete(id: string): Promise<boolean> {
    return this.users.delete(id);
  }
}
