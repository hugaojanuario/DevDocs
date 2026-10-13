import { defineConfig, type ProxyOptions } from 'vite';
import react from '@vitejs/plugin-react';
const target=process.env.DEVDOCS_PROXY_TARGET??'http://localhost:8080';
const localOrigins=['http://localhost:5173','http://127.0.0.1:5173','http://localhost:4173','http://127.0.0.1:4173'];
const proxy: Record<string, ProxyOptions>={
 '/api': {target,changeOrigin:true,configure(server){server.on('proxyReq',(request,source)=>{if(source.headers.origin && localOrigins.includes(source.headers.origin))request.setHeader('Origin',new URL(target).origin);});}},
 '/openapi.yaml': {target,changeOrigin:true},
};
export default defineConfig({plugins:[react()],server:{port:5173,strictPort:true,proxy},preview:{port:4173,strictPort:true,proxy},build:{sourcemap:false}});
