import { createContext, useContext, useState } from "react";
import React from "react";
import axios from "axios";

async function getUserByAppleAccountID(appleAccountID: string) {
    const url = process.env.EXPO_PUBLIC_API_URL + "/auth/login/apple";
    const response = await axios.post(url, {
        apple_id: appleAccountID,
    });
    const access_token = response.headers["access_token"];
    const refresh_token = response.headers["refresh_token"];

    console.log(access_token);
    console.log(refresh_token);

    axios.defaults.headers.common["Authorization"] = "Bearer " + access_token;
    axios.defaults.headers.common["refresh_token"] = refresh_token;

    const user = response.data;
    return user;
}

interface AuthContextType {
    user: any | null;
    login: (appleAccountID: string) => void;
    register: (firstName: string, lastName: string, email: string, appleAccountID: string) => any;
    logout: () => void;
    refresh: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<any | null>(null);

    async function register(firstName: string, lastName: string, email: string, appleAccountID: string) {
        const url = process.env.EXPO_PUBLIC_API_URL;
        console.log(url);
        try {
            const response = await fetch(url, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    appleAccountID,
                }),
            });

            if (!response.ok) {
                throw Error("Unable to complete operation" + " status code: " + response.statusText);
            }

            console.log(response);
            return response;
        } catch (e: any) {
            console.log(e);
        }
    }

    async function login(appleAccountID: string) {
        console.log(appleAccountID);
        console.log("Logging in...");
        const userRes = await getUserByAppleAccountID(appleAccountID);

        if (userRes) {
            setUser(userRes);
            return userRes;
        } else {
            alert("Could not login");
            throw new Error("Could not login");
        }
    }

    async function logout() {
        setUser(null);
    }

    async function refresh() {
        if (user) {
            login(user.appleAccountID);
        }
    }
    return <AuthContext.Provider value={{ user, register, login, logout, refresh }}>{children}</AuthContext.Provider>;
}

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error("useAuth must be used within an AuthProvider");
    }
    return context;
};
