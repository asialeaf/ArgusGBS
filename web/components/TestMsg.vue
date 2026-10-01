
<template>
<div class="msg-block">
    <div class="msg-header" v-if="work">
        信令交互 - {{topic}}
        <i class="fa fa-spinner fa-spin" v-show="live"></i>
        <a :href="url" class="msg-btn pull-right" v-show="!live && url && msg"><i class="fa fa-download"></i> 下载</a>
        <a role="button" class="msg-btn pull-right" v-show="!live && msg" v-clipboard="msg" @success="$message({type:'success', message:'成功拷贝到粘贴板'})"><i class="fa fa-copy"></i> 复制</a>
        <a role="button" class="msg-btn pull-right" v-show="!live && msg" @click="$store.dispatch('clear')"><i class="fa fa-trash-o"></i> 清空</a>
        <a role="button" class="msg-btn pull-right" v-show="live" @click="$store.dispatch('disconnect')"><i class="fa fa-stop"></i> 停止</a>
    </div>
    <div class="msg-header" v-else>
        <i class="fa fa-warning"></i> 服务器上没有安装 npcap,
        <a href="//npcap.com/#download" target="_blank">
            前往下载安装 <svg xmlns="http://www.w3.org/2000/svg" aria-hidden="true" x="0px" y="0px" viewBox="0 0 100 100" width="15" height="15" class="icon outbound"><path fill="currentColor" d="M18.8,85.1h56l0,0c2.2,0,4-1.8,4-4v-32h-8v28h-48v-48h28v-8h-32l0,0c-2.2,0-4,1.8-4,4v56C14.8,83.3,16.6,85.1,18.8,85.1z"></path> <polygon fill="currentColor" points="45.7,48.7 51.3,54.3 77.2,28.5 77.2,37.2 85.2,37.2 85.2,14.9 62.8,14.9 62.8,22.9 71.5,22.9"></polygon></svg>
        </a>
    </div>
    <div class="msg-content" :style="`height:${this.pageHeight}px;min-height:500px;`">
        <highlightjs :language="lang" :code="msg"/>
    </div>
</div>
</template>

<script>
export default {
    props: {
        lang: {
            type: String,
            default: 'sip',
        },
        topic: {
            type: String,
            default: '提示',
        },
        msg: {
            type: String,
            default: '',
        },
        work: {
            type: Boolean,
            default: true,
        },
        live: {
            type: Boolean,
            default: false,
        },
        url: {
            type: String,
            default: '',
        }
    },
    data() {
        return {
            pageHeight: 0,
            autoScroll: true,
            autoScrolling: false,
        }
    },
    watch: {
        msg: function(newVal, oldVal) {
            if(!newVal && oldVal) {
                this.autoScroll = true;
                return;
            }
            this.$nextTick(() => {
                var el = this.$el.querySelector(".msg-content");
                if(el && this.autoScroll) {
                    this.autoScrolling = true;
                    $(el).animate({
                        scrollTop: el.scrollHeight,
                    }, 100, () => {
                        this.autoScrolling = false;
                    });
                }
            });
        }
    },
    created() {
        this.initHeight();
    },
    mounted() {
        $(window).on('resize', this.initHeight);
        $(this.$el.querySelector(".msg-content")).on('scroll', this.onScroll);
    },
    beforeDestroy() {
        $(window).off('resize', this.initHeight);
        $(this.$el.querySelector(".msg-content")).off('scroll', this.onScroll);
    },
    methods: {
        initHeight() {
            this.pageHeight = window.innerHeight;
            if (typeof this.pageWidth != "number") {
                if (document.compatMode == "CSS1Compat") {
                    this.pageHeight = document.documentElement.clientHeight;
                } else {
                    this.pageHeight = document.body.clientHeight;
                }
            }
            this.pageHeight -= 160;
        },
        onScroll() {
            if(!this.autoScrolling) {
                var el = this.$el.querySelector(".msg-content");
                this.autoScroll = Math.abs(el.scrollTop + el.clientHeight - el.scrollHeight) < 200;
            }
        }
    }
}
</script>

<style lang="less" scoped>
@import url(~assets/styles/variables.less);

.msg-block {
  position: relative;
  overflow: hidden;
  margin: 0;
}

.msg-header {
    padding: 8px 12px;
    background: darken(@base, 10%);
    color: #eee;
    font-size: 14px;

    i.fa {
        padding: 4px;
        transition: color 0.2s;
    }

    .msg-btn, a {
        background: transparent;
        border: none;
        color: #ccc;
        cursor: pointer;
        margin-left: 10px;
    }

    .msg-btn:hover, a:hover {
        color: #fff;
    }

    .stop-btn {
        background: transparent;
        border: none;
        color: #e74c3c;
        cursor: pointer;
        margin-left: 10px;
    }

    .stop-btn:hover {
        color: lighten(#e74c3c, 10%);;
    }
}

.msg-content {
    position: relative;
    background: #282c34;
    overflow-x: auto;
    height: 100%;
}

pre {
    background: transparent;
    padding: 0;
    margin: 0;
    border: 0;
    border-radius: 0;
}
</style>
