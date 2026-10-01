<template>
<section class="content">
    <div class="row">
        <div class="col-sm-6">
            <form class="form-horizontal" autocomplete="off" ref="dev-form" spellcheck="false">
                <div class="form-group">
                    <label for="input-serial" class="col-sm-4 control-label">受测设备
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <el-select filterable clearable remote :remote-method="loadDevice" :loading="deviceLoading" :value="serial" :disabled="!!ws"
                            @change="changeDevice" @clear="clearDevice" @focus="!serial && loadDevice();" ref="selectDevice"
                            id="input-serial" name="serial" placeholder="可搜索" size="small" style="width:100%;">
                            <el-option v-for="item in devices" :key="item.ID" :label="item.ID" :value="item.ID">
                            </el-option>
                        </el-select>
                    </div>
                </div>
                <div class="form-group" v-show="['目录查询', '注册心跳', '远程启动', '软件升级'].indexOf(topic) < 0">
                    <label for="input-code" class="col-sm-4 control-label">{{ decode ? '监视器ID' : '受测通道' }}
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <el-select filterable clearable remote :remote-method="loadChannel" :loading="channelLoading" :value="code" :disabled="!!ws"
                            @change="changeChannel"  @clear="updateCode('');" @focus="!code && loadChannel();" ref="selectChannel"
                            id="input-code" name="code" placeholder="可搜索" size="small" style="width:100%;">
                            <el-option v-for="item in channels" :key="item.ID" :label="item.ID" :value="item.ID">
                            </el-option>
                        </el-select>
                    </div>
                </div>
            </form>
            <router-view @show-result="showResut"></router-view>
            <div class="clearfix"></div>
            <div class="panel panel-success panel-result" v-if="result && resultVisible" style="margin-top: 50px;">
                <div class="panel-heading">结果展示</div>
                <div class="panel-body">
                    <json-viewer :value="result" :expand-depth=2 copyable boxed @copied="$message({type:'success', message:'成功拷贝到粘贴板'})">
                        <template slot="copy">
                            <span class="text-primary"><i class="fa fa-copy"></i> 复制</span>
                        </template>
                    </json-viewer>
                </div>
            </div>
        </div>
        <div class="col-sm-6">
            <TestMsg id="msg-sticky" :work="work" :topic="topic" :msg="msg" :live="!!ws" :url="url"></TestMsg>
        </div>
    </div>
</section>
</template>

<script>
import "assets/js/jquery-sticky.js";
import "jquery-slimscroll";
import { mapState, mapActions, mapMutations } from "vuex";
import TestMsg from 'components/TestMsg.vue'

export default {
    data() {
        return {
            deviceLoading: false,
            channelLoading: false,
            resultVisible: true,
            devices: [],
            channels: [],
        }
    },
    computed: {
        ...mapState(["work", "msg", "ws", "topic", "url", "serial", "code", "result"]),
        decode() {
            if(this.serial && this.serial.length >= 13 && this.serial.substr(10, 3) == "114") {
                return true;
            }
            if(this.code && this.code.length >= 13 && this.code.substr(10, 3) == "133") {
                return true;
            }
            return false;
        },
    },
    components: { TestMsg },
    methods: {
        ...mapActions(["check", "connect", "disconnect", "clear"]),
        ...mapMutations(["updateSerial", "updateCode"]),
        loadDevice(q = "") {
            this.deviceLoading = true;
            $.get("/api/v1/device/list", {
                q: q == this.serial ? '' : q,
                limit: 10,
            }).then(ret => {
                this.devices = ret.DeviceList || [];
                // if(!this.serial && !q && this.devices.length > 0) {
                //     this.updateSerial(this.devices[0].ID);
                // }
            }).always(() => {
                this.deviceLoading = false;
            })
        },
        changeDevice(v = "") {
            this.updateSerial(v);
            this.updateCode("");
            this.loadChannel();
            this.clear();
        },
        clearDevice() {
            this.updateSerial("");
            this.updateCode("");
            this.clear();
        },
        loadChannel(q = "") {
            if(!this.serial) {
                this.channels.length = 0;
                return;
            }
            this.channelLoading = true;
            $.get("/api/v1/device/channellist", {
                q: q == this.code ? '' : q,
                serial: this.serial,
                limit: 10,
            }).then(ret => {
                this.channels = ret.ChannelList || [];
                if(this.serial && !this.code && !q && this.channels.length > 0) {
                    this.updateCode(this.channels[0].ID);
                }
            }).always(() => {
                this.channelLoading = false;
            })
        },
        changeChannel(v = "") {
            this.updateCode(v);
        },
		updateSticky() {
            $("#msg-sticky").sticky("update");
            this.$nextTick(() => {
                $("#msg-sticky").sticky("update");
            })
		},
        showResut(visible = true) {
            this.resultVisible = visible;
        },
    },
    async created() {
        var work = await this.check();
        if(!work) {
            console.log("work =", work);
        }
    },
    mounted() {
        $("#msg-sticky").sticky({ bottomSpacing: 40, topSpacing: 65});
        $("body").addClass("fixed");
        $("aside.main-sidebar").slimscroll({
            width: $("aside.main-sidebar").width(),
            height: '100%',
            color: '#20262d',
            size: '0',
        });
        $(document).on('expanded.pushMenu collapsed.pushMenu', this.updateSticky);
    },
    beforeDestroy() {
        $(document).off('expanded.pushMenu collapsed.pushMenu', this.updateSticky);
        $("#msg-sticky").unstick();
    },
	beforeRouteUpdate(to, from, next) {
		this.clear();
        this.resultVisible = true;
		next();
	}
}
</script>

<style lang="less">
.hljs-keyword {
    color: #9952e1;
    font-weight: bold;
}

.hljs-literal {
    color: #859900;
    font-weight: bold;
}

.hljs-success {
    color: #2ecc71;
    font-weight: bold;
}

.hljs-info {
    color: #61aeee;
    font-weight: bold;
}

.hljs-warn {
    color: #c70;
    font-weight: bold;
}

.hljs-error {
    color: #e74c3c;
    font-weight: bold;
}

.el-slider__runway, .el-progress-bar__outer {
    background-color: #ddd;
}
</style>
