import plugin from 'cypress-plugin-retries/lib/plugin'
import fs from 'fs'
import http from 'http'

function makeDoCall(path){
    return () => new Promise((resolve, reject)=>{
    http.get('http://127.0.0.1:3033'+path, res=>{
        if (res.statusCode !== 200) {
            reject(new Error("request failed: "+res.statusCode))
            return
        }
        resolve()
    })
})
}

export const tasks = {
	'engine:trigger': makeDoCall('/signal?sig=SIGUSR2'),
	'engine:start': makeDoCall('/start'),
	'engine:stop': makeDoCall('/stop'),
}

export default on => {
	plugin(on)
	on('task', tasks)
}