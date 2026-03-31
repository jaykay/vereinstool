import { writable } from 'svelte/store';
import { api, ApiError } from './api';

export interface User {
	id: number;
	email: string;
	name: string;
	role: 'admin' | 'member';
}

function createUserStore() {
	const { subscribe, set } = writable<User | null>(null);

	return {
		subscribe,
		set,
		async load() {
			try {
				const data = await api.get<{ user: User }>('/auth/me');
				set(data.user);
				return data.user;
			} catch (e) {
				if (e instanceof ApiError && e.status === 401) {
					set(null);
					return null;
				}
				throw e;
			}
		},
		async login(email: string, password: string) {
			const data = await api.post<{ user: User }>('/auth/login', { email, password });
			set(data.user);
			return data.user;
		},
		async logout() {
			await api.post('/auth/logout');
			set(null);
		},
		clear() {
			set(null);
		}
	};
}

export const user = createUserStore();
