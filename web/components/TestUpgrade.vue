<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">设备厂商
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" v-model="manufacturer" class="form-control">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">固件版本
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" v-model="firmware" class="form-control">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">下载链接
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <div class="input-group">
                    <input type="text" class="form-control" spellcheck="false" autocomplete="off" v-model.trim="fileurl">
                    <span class="input-group-addon" title="上传中..." v-if="uploading">
                        <i class="fa fa-spin fa-spinner"></i>
                    </span>
                    <el-upload action="/upgrade" accept=".bin,.dav,.img,.dig,.rom,.pak" :show-file-list="false" class="input-group-btn"
                        :on-success="onUploadSuccess" :on-error="onUploadError" :on-progress="onUploadProgress" v-else>
                        <button type="button" class="btn btn-default" style="border-top-right-radius:0;border-bottom-right-radius:0;border-top-left-radius:0;border-bottom-left-radius:0;border-left:0;" title="上传Cert证书">
                            <i class="fa fa-upload"></i>
                        </button>
                    </el-upload>
                </div>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">会话ID
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" v-model="session_id" class="form-control">
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending||uploading">发送<span v-show="sending">...</span></button>
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
const regUpgradeResult = /<UpgradeResult>(.+)<\/UpgradeResult>/m;
export default {
    data() {
        return {
            manufacturer: "",
            firmware: "",
            fileurl: "",
            session_id: "",
            sending: false,
            uploading: false,
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
                match = regUpgradeResult.exec(last);
                if(match && match.length > 1 && match[1]) {
                    this.pushResult(`升级结果: ${match[1]}`);
                    setTimeout(() => {
                        this.$store.dispatch("disconnect");
                    }, 1000);
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

            $.get("/api/v1/control/deviceupgrade", {
                serial: this.$store.state.serial,
                code: this.$store.state.serial,
                manufacturer: this.manufacturer,
                firmware: this.firmware,
                fileurl: this.fileurl,
                session_id: this.session_id,
                timeout: 0,
            }).then(ret => {
                this.pushResult("软件升级下发成功");
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
        onUploadSuccess(res, file, fileList) {
            this.uploading = false;
            this.fileurl = `${res.FileURL}/${res.FileMD5}`;
            this.session_id = res.FileMD5;
            this.$message({
                type: "success",
                message: "上传固件包成功！"
            });
        },
        onUploadError(err, file, fileList) {
            this.uploading = false;
            if (err) {
                this.$message({
                    type: "error",
                    message: err + ""
                })
            }
        },
        onUploadProgress(event, file, fileList) {
            this.uploading = true;
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