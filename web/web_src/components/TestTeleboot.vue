<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-danger" @click.prevent="send" :disabled="sending">远程启动<span v-show="sending">...</span></button>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import { mapState } from "vuex";
const regKeepalive = /^Call-ID:\s+(.+)$[\s\S]+<CmdType>Keepalive<\/CmdType>/m;
const regRegister = /^REGISTER\s+[\s\S]+^Call-ID:\s+(.+)$/m;
const regExpires = /^Expires:\s+(.*)$/m;
const regResponse = /^SIP\/2\.0\s+(.+)$[\s\S]+Call-ID:\s+(.+)$/m;
export default {
    data() {
        return {
            sending: false,
            callMap: {},
            queue: [],
            result: [],
            timer: 0,
        }
    },
    computed: {
        ...mapState(["msg"]),
    },
    watch: {
        msg: function(newVal, oldVal) {
            if(!newVal || newVal.length <= oldVal.length) return;
            var last = newVal.substr(oldVal.length);
            if(last.startsWith("<<<<<<")) {
                var match = regKeepalive.exec(last);
                if(match && match.length > 1 && match[1]) {
                    this.callMap[match[1]] = "心跳";
                    this.pushResult(`收到心跳`);
                    return
                }
                match = regRegister.exec(last);
                if(match && match.length > 1 && match[1]) {
                    var callid = match[1];
                    match = regExpires.exec(last)
                    if(match && match.length > 1) {
                        var exp = parseInt(match[1]);
                        if(!exp) {
                            this.callMap[callid] = "注销";
                            this.pushResult(`收到注销`);
                            return
                        }
                    }
                    this.callMap[callid] = "注册";
                    this.pushResult(`收到注册`);
                    return
                }
            } else if(Object.keys(this.callMap).length) {
                var match = regResponse.exec(last);
                if(match && match.length > 2 && this.callMap[match[2]]) {
                    this.pushResult(`回复${this.callMap[match[2]]} ${match[1]}`);
                    delete this.callMap[match[2]];
                }
            }
        },
    },
    methods: {
        async send() {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.get("/api/v1/control/teleboot", {
                serial: this.$store.state.serial,
                code: this.$store.state.serial,
            }).then(ret => {
                this.pushResult("远程启动");
                // setTimeout(() => {
                //     this.$store.dispatch("disconnect");
                // }, 1000);
            }).always(() => {
                this.sending = false;
            })
        },
        pushResult(result) {
            if(!result) return;
            this.queue.push(result);
            if(this.timer) {
                clearTimeout(this.timer);
                this.timer = 0;
            }
            this.timer = setTimeout(() => {
                this.result.push(...this.queue);
                this.queue.length = 0;
                if(!this.$store.state.result) {
                    if(this.$store.state.ws) {
                        this.$store.commit("updateResult", this.result);
                    } else {
                        this.result.length = 0;
                    }
                }
            }, 1000);
        },
        reset() {
            if(this.timer) {
                clearTimeout(this.timer);
                this.timer = 0;
            }
        },
        beforeUnload(event) {
            this.reset();
            event.returnValue = '';
        },
    },
    mounted() {
        $(window).on("beforeunload", this.beforeUnload);
    },
    beforeDestroy() {
        $(window).off("beforeunload", this.beforeUnload);
        this.reset();
    },
}
</script>
