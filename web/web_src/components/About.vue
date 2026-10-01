<template>
<div class="container-fluid no-padding box-cards">
    <el-card class="box-card col-lg-6" shadow="never">
        <div slot="header" class="clearfix">
            <div class="col-md-12 no-padding">
                <h3>信令服务</h3>
            </div>
        </div>
        <div class="server-info">
            <div class="box box-widget">
                <div class="box-header">
                    <h4> <i class="fa fa-support"></i> 版本信息</h4>
                </div>
                <div class="box-body table-responsive no-padding">
                    <table class="table table-striped">
                        <tbody>
                            <tr>
                                <td style="width:20%;">硬件信息</td>
                                <td><span>{{serverInfo.Hardware}}</span></td>
                            </tr>
                            <tr>
                                <td>接口版本</td>
                                <td><span id="interface-info">{{serverInfo.InterfaceVersion}}</span></td>
                            </tr>
                            <tr>
                                <td>运行时间</td>
                                <td>
                                    <span id="running-time-info">{{runningTime || serverInfo.RunningTime}}
                                        <small v-if="userInfo">
                                            &nbsp;<a href="#" @click.prevent="restart" class="text-orange">重启</a>
                                        </small>
                                    </span>
                                </td>
                            </tr>
                            <tr>
                                <td>软件信息</td>
                                <td><span id="software-info">{{serverInfo.LogoText === 'LiveGBS' ? serverInfo.Server : (serverInfo.Server||"").replace("LiveCMS", "CMS")}}</span></td>
                            </tr>
                            <tr v-if="!isDemoUser(serverInfo, userInfo) && (debug || sharking || smsshark.output)">
                                <td>网络抓包</td>
                                <td>
                                    <el-input placeholder="过滤" v-model.trim="filter" size="small" style="margin-bottom:10px;" :disabled="sharking" clearable>
                                        <el-select v-model.trim="iface" slot="prepend" placeholder="选择网卡" style="width:100px;" :disabled="sharking">
                                            <el-option label="所有网卡" value="any"></el-option>
                                            <el-option :label="node.Name" :value="node.Name" v-for="(node, idx) in ifaces" :key="idx">
                                                <span style="float: left">{{ node.Name }}</span>
                                                <span style="float: right; color: #8492a6; font-size: 13px">&nbsp;{{ node.IP }}</span>
                                            </el-option>
                                        </el-select>
                                    </el-input>
                                    <div>
                                        <button type="button" class="btn btn-sm btn-primary" @click.prevent="sharkStart" v-if="!sharking" :disabled="bSubmitting">开始</button>
                                        <button type="button" class="btn btn-sm btn-danger" @click.prevent="sharkStop" v-if="sharking" :disabled="bSubmitting">停止</button>
                                        <span>&nbsp;&nbsp;</span>
                                        <span v-if="sharking && sharkSeconds">
                                            时长({{sharkSeconds}}秒) 数量({{sharkCount}}) 大小({{sharkBytes}})
                                        </span>
                                    </div>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
            <div class="box box-widget">
                <div class="box-header">
                    <h4>
                        <i class="fa" :class="{'fa-key': !dongle, 'fa-usb': dongle}"></i> 授权信息 <span v-if="serverInfo.VersionType">({{serverInfo.VersionType}})</span>
                        &nbsp;&nbsp;
                        <small v-if="canExpand">
                            <a href="#" @click.prevent="toggleExpand" :class="{ 'fa':true,'fa-plus':!expanded,'fa-chevron-down':expanded}"
                                :title="expanded ? '关闭激活码入口' : '显示激活码入口'" style="color:#afa9a9;"></a>
                        </small>
                    </h4>
                </div>
                <div class="box-body table-responsive no-padding">
                    <table class="table table-striped">
                        <tbody>
                            <tr>
                                <td style="width:20%;">授权对象</td>
                                <td>{{serverInfo.Authorization}}</td>
                            </tr>
                            <tr>
                                <td>授权时间</td>
                                <td v-if="!actived">
                                    <span v-if="!changed || serverInfo.RemainDays > 0">剩余期限{{serverInfo.RemainDays}}天</span>
                                    <span v-if="changed" style="color:red;">硬件信息变动 {{changed}}</span>
                                </td>
                                <td v-else>永久授权</td>
                            </tr>
                            <tr v-if="(actived || expanded || serverInfo.RemainDays > 30) && serverInfo.ChannelCount >= 0 && serverInfo.ChannelCount != 1000">
                                <td>通道数</td>
                                <td>{{serverInfo.ChannelCount}}</td>
                            </tr>
                            <tr v-if="!actived || expanded">
                                <td>机器码</td>
                                <td>
                                    <span id="key-info">CMS{{(serverInfo.Server||"").indexOf("Linux") >= 0 ? "L":"W"}}{{requestkey}}</span>
                                    <br v-if="serverInfo.LogoText === 'LiveGBS'">
                                    <span v-if="serverInfo.LogoText === 'LiveGBS'" style="color:#bbb">(通过邮箱：support@liveqing.com 向商务人员咨询永久授权信息)</span>
                                </td>
                            </tr>
                            <tr v-if="!actived || expanded">
                                <td>提交激活码</td>
                                <td>
                                    <el-input type="textarea" :rows="1" placeholder="输入申请到的激活码" v-model.trim="activationCode" ref="activationCode" :autosize="{minRows:1, maxRows:5}" style="margin-bottom:10px;"></el-input>
                                    <div>
                                        <button type="button" class="btn btn-sm btn-primary" @click.prevent="checkCode" :disabled="bSubmitting || !activationCode"> 提交 </button>
                                    </div>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </el-card>
    <el-card class="box-card col-lg-6" shadow="never">
        <div slot="header" class="clearfix">
            <div class="col-md-6 no-padding">
                <h3>流媒体服务</h3>
            </div>
            <div class="col-md-6 no-padding" v-if="sms.Load">
                <select style="margin-top: 23px;width:100%" v-model.trim="smsserial" @change="smschange">
                    <option v-for="(c, idx) in smss" :value="c.Serial" :key="idx"> SMS-{{c.Serial}}</option>
                </select>
            </div>
        </div>
        <div class="server-info">
            <h4 style="text-align:center;" v-if="!smsserverinfo.Hardware"> SMS 流媒体服务尚未启动 </h4>
            <div class="box box-widget" v-if="smsserverinfo.Hardware">
                <div class="box-header">
                    <h4> <i class="fa fa-support"></i> 版本信息</h4>
                </div>
                <div class="box-body table-responsive no-padding">
                    <table class="table table-striped">
                        <tbody>
                            <tr>
                                <td style="width:20%;">硬件信息</td>
                                <td><span>{{smsserverinfo.Hardware}}</span></td>
                            </tr>
                            <tr>
                                <td style="width:20%;">负载</td>
                                <td>
                                    <span v-if="sms.Load > 0 && !isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                                        <a @click.prevent="$emit('show-session-list', '', sms.Serial)" role="button" class="text-primary">{{sms.Load}}</a>
                                    </span>
                                    <span v-else>{{sms.Load}}</span>
                                </td>
                            </tr>
                            <tr v-if="(sms.H264Decoder || sms.HevcDecoder) && sms.HWDecodeLoad">
                                <td style="width:20%;">硬解负载</td>
                                <td><span>{{sms.HWDecodeLoad}}</span></td>
                            </tr>
                            <tr v-if="(sms.H264Encoder || sms.HevcEncoder) && sms.HWEncodeLoad">
                                <td style="width:20%;">硬编负载</td>
                                <td><span>{{sms.HWEncodeLoad}}</span></td>
                            </tr>
                            <tr>
                                <td>运行时间</td>
                                <td>
                                    <span id="running-time-info">{{runningSMSTime || smsserverinfo.RunningTime}}
                                        <small v-if="userInfo">
                                            &nbsp;<a href="#" @click.prevent="restartSMS" class="text-orange">重启</a>
                                        </small>
                                    </span>
                                </td>
                            </tr>
                            <tr>
                                <td>软件信息</td>
                                <td><span id="software-info">{{serverInfo.LogoText === 'LiveGBS' ? smsserverinfo.Server : (smsserverinfo.Server||"").replace("LiveSMS", "SMS")}}</span></td>
                            </tr>
                            <tr v-if="!isDemoUser(serverInfo, userInfo) && (debug || sharking || smsshark.output)">
                                <td>网络抓包</td>
                                <td>
                                    <el-input placeholder="过滤" v-model.trim="smsshark.filter" size="small" style="margin-bottom:10px;" :disabled="!!smsshark.output" clearable>
                                        <el-select v-model.trim="smsshark.iface" slot="prepend" placeholder="选择网卡" style="width:100px;" :disabled="!!smsshark.output">
                                            <el-option label="所有网卡" value="any"></el-option>
                                            <el-option :label="node.Name" :value="node.Name" v-for="(node, idx) in smsshark.ifaces" :key="idx">
                                                <span style="float: left">{{ node.Name }}</span>
                                                <span style="float: right; color: #8492a6; font-size: 13px">&nbsp;{{ node.IP }}</span>
                                            </el-option>
                                        </el-select>
                                    </el-input>
                                    <div>
                                        <button type="button" class="btn btn-sm btn-primary" @click.prevent="sharkStartSMS" v-if="!smsshark.output" :disabled="bSubmittingSMS">开始</button>
                                        <button type="button" class="btn btn-sm btn-danger" @click.prevent="sharkStopSMS" v-if="smsshark.output" :disabled="bSubmittingSMS">停止</button>
                                        <span>&nbsp;&nbsp;</span>
                                        <span v-if="smsshark.output && smsshark.seconds">
                                            时长({{smsshark.seconds}}秒) 数量({{smsshark.count}}) 大小({{smsshark.bytes}})
                                        </span>
                                    </div>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
            <div class="box box-widget" v-if="smsrequestkey.RequestKey && smsserverinfo.Authorization">
                <div class="box-header">
                    <h4> <i class="fa" :class="{'fa-key': !smsrequestkey.Dongle, 'fa-usb': !!smsrequestkey.Dongle}"></i> 授权信息 <span v-if="smsserverinfo.VersionType">({{smsserverinfo.VersionType}})</span></h4>
                </div>
                <div class="box-body table-responsive no-padding">
                    <table class="table table-striped">
                        <tbody>
                            <tr>
                                <td style="width:20%;">授权对象</td>
                                <td>{{smsserverinfo.Authorization}}</td>
                            </tr>
                            <tr>
                                <td>授权时间</td>
                                <td v-if="!activedsms">
                                    <span v-if="!smsrequestkey.Changed || smsserverinfo.RemainDays > 0">剩余期限{{smsserverinfo.RemainDays}}天</span>
                                    <span v-if="smsrequestkey.Changed" style="color:red;">硬件信息变动 {{smsrequestkey.Changed}}</span>
                                </td>
                                <td v-else>永久授权</td>
                            </tr>
                            <tr v-if="(activedsms || expanded) && smsserverinfo.ChannelCount && smsserverinfo.ChannelCount >= 0 && smsserverinfo.ChannelCount != 1000">
                                <td>通道数</td>
                                <td>{{smsserverinfo.ChannelCount}}</td>
                            </tr>
                            <tr v-if="!activedsms || expanded">
                                <td>机器码</td>
                                <td>
                                    <span id="key-info">SMS{{(smsserverinfo.Server||"").indexOf("Linux") >= 0 ? "L":"W"}}{{smsrequestkey.RequestKey}}</span>
                                    <br v-if="serverInfo.LogoText === 'LiveGBS'">
                                    <span v-if="serverInfo.LogoText === 'LiveGBS'" style="color:#bbb">(通过邮箱：support@liveqing.com 向商务人员咨询永久授权信息)</span>
                                </td>
                            </tr>
                            <tr v-if="!activedsms || expanded">
                                <td>提交激活码</td>
                                <td>
                                    <el-input type="textarea" :rows="1" placeholder="输入申请到的激活码" v-model.trim="activationCodeSMS" ref="activationCodeSMS" :autosize="{minRows:1, maxRows:5}" style="margin-bottom:10px;"></el-input>
                                    <div>
                                        <button type="button" class="btn btn-sm btn-primary" @click.prevent="checkCodeSMS" :disabled="bSubmittingSMS || !activationCodeSMS"> 提交 </button>
                                    </div>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </el-card>
</div>
</template>

<script>
import { mapState, mapActions } from "vuex";
import moment from 'moment';
export default {
    data() {
        return {
            timer: 0,
            requestkey: "",
            activationCode: "",
            activationCodeSMS: "",
            bSubmitting: false,
            bSubmittingSMS: false,
            runningTime: "",
            runningSMSTime: "",
            smsserial: "",
            smss: [],
            sms: {},
            smsrequestkey: {},
            smsserverinfo: {},
            smsshark: {
                serial: "",
                output: "",
                count: 0,
                seconds: 0,
                bytes: "0B",
                iface: "any",
                ifaces: [],
                filter: "",
            },
            expanded: false,
            dongle: false,
            changed: "",
            debug: false,
            sharking: false,
            sharkCount: 0,
            sharkSeconds: 0,
            sharkBytes: "0B",
            iface: "any",
            ifaces: [],
            filter: "",
        };
    },
    computed: {
        ...mapState(["userInfo", "serverInfo"]),
        actived() {
            return this.serverInfo.RemainDays == 9999;
        },
        activedsms() {
            return this.smsserverinfo.RemainDays == 9999;
        },
        canExpand() {
            if (this.actived && this.serverInfo.VersionType && this.serverInfo.VersionType.indexOf("旗舰版") < 0) {
                return true;
            }
            if (this.actived && this.serverInfo.ChannelCount && this.serverInfo.ChannelCount >= 0 && this.serverInfo.ChannelCount != 1000) {
                return true;
            }
            return false;
        },
    },
    created() {
        this.expanded = this.$getQueryString("expand") == "yes";
        this.debug = this.$getQueryString("debug") == "yes";
    },
    mounted() {
        // this.getServerInfo();
        this.timer = setInterval(() => {
            if (this.serverInfo && this.serverInfo.StartUpTime) {
                var start = moment(this.serverInfo.StartUpTime, "YYYY-MM-DD HH:mm:ss");
                var now = moment();
                var d = moment.duration(now.diff(start));
                if (this.serverInfo.DiffDuration) {
                    d = d.add(this.serverInfo.DiffDuration);
                }
                this.runningTime = `${parseInt(d.asDays())} Days ${d.hours()} Hours ${d.minutes()} Mins ${d.seconds()} Secs`;
            }
            if (this.smsserverinfo && this.smsserverinfo.StartUpTime) {
                var start = moment(this.smsserverinfo.StartUpTime, "YYYY-MM-DD HH:mm:ss");
                var now = moment();
                var d = moment.duration(now.diff(start));
                if (this.smsserverinfo.DiffDuration) {
                    d = d.add(this.smsserverinfo.DiffDuration);
                }
                this.runningSMSTime = `${parseInt(d.asDays())} Days ${d.hours()} Hours ${d.minutes()} Mins ${d.seconds()} Secs`;
            }
            if (this.sharking) {
                $.ajax({
                    method: "GET",
                    url: "/api/v1/shark/stats",
                    global: false,
                }).then(ret => {
                    this.sharking = !!ret["Sharking"];
                    this.sharkCount = ret["Count"] || 0;
                    this.sharkSeconds = ret["Seconds"] || 0;
                    this.sharkBytes = ret["HBytes"] || "0B";
                })
            }
            if (this.smsshark.serial && this.smsshark.output) {
                $.ajax({
                    method: "GET",
                    url: `/sms/${this.smsshark.serial}/api/v1/shark/stats/${this.smsshark.output}`,
                    global: false,
                }).then(ret => {
                    this.smsshark.count = ret["Count"] || 0;
                    this.smsshark.seconds = ret["Seconds"] || 0;
                    this.smsshark.bytes = ret["HBytes"] || "0B";
                }).fail(xhr => {
                    this.smsshark.output = "";
                    this.smsshark.count = 0;
                    this.smsshark.seconds = 0;
                    this.smsshark.bytes = "0B";
                })
            }
        }, 1000)
        $.get("/api/v1/getrequestkey").then(ret => {
            this.requestkey = ret.RequestKey;
            this.dongle = !!ret.Dongle;
            this.changed = ret.Changed||"";
        });
        this.sharking = !!this.serverInfo.Sharking;
        this.getSMSList();
        $(document).on("keydown", this.keyDown);
    },
    beforeDestroy() {
        if (this.timer) {
            clearInterval(this.timer);
            this.timer = 0;
        }
        $(document).off("keydown", this.keyDown);
    },
    beforeRouteEnter(to, from, next) {
        next(vm => {
            if(to.query.expand === "yes") {
                vm.expanded = true;
            }
            if(to.query.debug === "yes") {
                vm.debug = true;
            }
            if(vm.debug || vm.sharking) {
                vm.getSharkDevs();
            }
        });
    },
    methods: {
        ...mapActions([ "getServerInfo" ]),
        restart() {
            this.$confirm('此操作将重启信令服务, 是否继续?', '提示', {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/restart").then(ret => {
                    this.$message({
                        type: 'success',
                        message: '重启成功!'
                    });
                    setTimeout(() => {
                        this.getServerInfo();
                    }, 2000);
                }).fail(() => {
                    this.$message({
                        type: 'error',
                        message: '重启失败!'
                    });
                })
            }).catch(() => {});
        },
        restartSMS() {
            this.$confirm('此操作将重启流媒体服务, 是否继续?', '提示', {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/sms/restart", {
                    serial: this.smsserial
                }).then(ret => {
                    this.$message({
                        type: 'success',
                        message: '重启成功!'
                    });
                    setTimeout(() => {
                        this.getSMSInfo();
                    }, 2000);
                }).fail(() => {
                    this.$message({
                        type: 'error',
                        message: '重启失败!'
                    });
                })
            }).catch(() => {});
        },
        checkCode() {
            if (this.activationCode == "") {
                this.$message({
                    type: "error",
                    message: "请输入激活码"
                });
            } else {
                this.bSubmitting = true;
                $.post("/api/v1/verifyproductcode", {
                    productcode: this.activationCode
                }).then(ret => {
                    if (ret.State == 1) {
                        this.$message({
                            type: "success",
                            message: "授权成功！"
                        });
                        this.getServerInfo();
                    } else {
                        this.$message({
                            type: "error",
                            message: "输入有效激活码"
                        })
                    }
                }).always(() => {
                    this.activationCode = "";
                    this.bSubmitting = false;
                })
            }
        },
        getSharkDevs() {
            $.ajax({
                method: "GET",
                url: "/api/v1/shark/devices",
                global: false,
            }).then(ret => {
                this.ifaces = ret["List"] || [];
            })
        },
        getSMSSharkDevs() {
            if(this.smsserial) {
                $.ajax({
                    method: "GET",
                    url: `/sms/${this.smsserial}/api/v1/shark/devices`,
                    global: false,
                }).then(ret => {
                    this.smsshark.ifaces = ret["List"] || [];
                })
            }
        },
        sharkStart() {
            this.bSubmitting = true;
            $.ajax({
                method: "POST",
                url: "/api/v1/shark/start",
                global: false,
                data: {
                    iface: this.iface,
                    filter: this.filter,
                }
            }).then(ret => {
                this.sharking = true;
            }).fail(xhr => {
                var msg = "操作失败";
                if(xhr && xhr.responseText) {
                    msg = xhr.responseText;
                }
                try {
                    msg = JSON.parse(msg);
                } catch (error) {}
                if(msg === "npcap not found") {
                    this.$confirm("服务器上没有安装 npcap, 是否前往下载安装?", "提示", {
                        confirmButtonText: '确定',
                        cancelButtonText: '取消',
                        lockScroll: false,
                    }).then(() => {
                        window.open("//npcap.com/#download", "_blank");
                    }).catch(() => {});
                } else {
                    this.$message({
                        type: 'error',
                        message: msg
                    })
                }
            }).always(() => {
                this.bSubmitting = false;
            })
        },
        sharkStop() {
            this.bSubmitting = true;
            $.post("/api/v1/shark/stop").then(ret => {
                this.sharking = false;
                this.sharkCount = 0;
                this.sharkSeconds = 0;
                this.sharkBytes = "0B";
                if(ret["Output"]) {
                    window.open(`/api/v1/shark/download/${ret["Output"]}`, "_blank");
                }
            }).always(() => {
                this.bSubmitting = false;
            })
        },
        sharkStartSMS() {
            this.bSubmittingSMS = true;
            $.ajax({
                method: "POST",
                url: "/api/v1/sms/shark/start",
                global: false,
                data: {
                    serial: this.smsserial,
                    iface: this.smsshark.iface,
                    filter: this.smsshark.filter,
                }
            }).then(ret => {
                this.smsshark.serial = this.smsserial;
                this.smsshark.output = ret["Output"] || "";
            }).fail(xhr => {
                var msg = "操作失败";
                if(xhr && xhr.responseText) {
                    msg = xhr.responseText;
                }
                try {
                    msg = JSON.parse(msg);
                } catch (error) {}
                if(msg === "npcap not found") {
                    this.$confirm("服务器上没有安装 npcap, 是否前往下载安装?", "提示", {
                        confirmButtonText: '确定',
                        cancelButtonText: '取消',
                        lockScroll: false,
                    }).then(() => {
                        window.open("//npcap.com/#download", "_blank");
                    }).catch(() => {});
                } else {
                    this.$message({
                        type: 'error',
                        message: msg
                    })
                }
            }).always(() => {
                this.bSubmittingSMS = false;
            })
        },
        sharkStopSMS() {
            this.bSubmittingSMS = true;
            $.post("/api/v1/sms/shark/stop", {
                serial: this.smsshark.serial,
            }).then(ret => {
                this.smsshark.output = "";
                this.smsshark.count = 0;
                this.smsshark.seconds = 0;
                this.smsshark.bytes = "0B";
                if(ret["Output"]) {
                    window.open(`/sms/${this.smsshark.serial}/api/v1/shark/download/${ret["Output"]}`, "_blank");
                }
            }).always(() => {
                this.bSubmittingSMS = false;
            })
        },
        keyDown(e) {
            if(e.altKey && e.shiftKey) {
                switch(e.key) {
                    case 'D':
                        e.preventDefault();
                        this.toggleDebug();
                        break;
                    case 'E':
                        e.preventDefault();
                        this.toggleExpand();
                        break;
                }
            }
        },
        toggleExpand() {
            this.expanded = !this.expanded;
        },
        toggleDebug() {
            this.debug = !this.debug;
            if(this.debug && (!this.ifaces || !this.ifaces.length)) {
                this.getSharkDevs();
            }
            if(this.debug && (!this.smsshark.ifaces || !this.smsshark.ifaces.length)) {
                this.getSMSSharkDevs();
            }
        },
        getSMSList() {
            if (this.smsserial == "") {
                $.get("/api/v1/sms/list").then(ret => {
                    this.smss = ret;
                    if (ret.length > 0) {
                        this.sms = ret[0];
                        this.smsserial = ret[0].Serial;
                    }
                    this.getSMSInfo();
                })
            }
        },
        getSMSInfo() {
            if (this.smsserial != "") {
                $.get("/api/v1/sms/getrequestkey", {
                    serial: this.smsserial
                }).then(ret => {
                    this.smsrequestkey = ret;
                })
                $.get("/api/v1/sms/getserverinfo", {
                    serial: this.smsserial
                }).then(ret => {
                    this.smsserverinfo = ret;
                    this.smsshark.serial = this.smsserial;
                    this.smsshark.output = ret["SharkOutput"] || "";
                    if(this.debug || this.sharking || this.smsshark.output) {
                        this.getSMSSharkDevs();
                        if(!this.ifaces || !this.ifaces.length) {
                            this.getSharkDevs();
                        }
                    }
                })
            }
        },
        smschange() {
            this.getSMSInfo()
            $.get("/api/v1/sms/list").then(ret => {
                this.smss = ret;
                for (var i = 0; i < this.smss.length; i++) {
                    if (this.smss[i].Serial == this.smsserial) {
                        this.sms = this.smss[i];
                        break
                    }
                }
            })
        },
        checkCodeSMS() {
            if (this.activationCodeSMS == "") {
                this.$message({
                    type: "error",
                    message: "请输入激活码"
                });
            } else {
                this.bSubmittingSMS = true;
                $.post("/api/v1/sms/verifyproductcode", {
                    serial: this.smsserial,
                    productcode: this.activationCodeSMS
                }).then(ret => {
                    if (ret.State == 1) {
                        this.$message({
                            type: "success",
                            message: "授权成功！"
                        });
                        this.getSMSInfo();
                    } else {
                        this.$message({
                            type: "error",
                            message: "输入有效激活码"
                        })
                    }
                }).always(() => {
                    this.activationCodeSMS = "";
                    this.bSubmittingSMS = false;
                })
            }
        },
    }
}
</script>

<style lang="less" scoped>
.container-fluid.no-padding.box-cards {
    overflow: hidden;

    .box-card {
        &[class*="col-"] {
            margin-bottom: -99999px;
            padding-bottom: 99999px;
        }
    }
}
</style>
