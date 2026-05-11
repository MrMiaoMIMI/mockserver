<template>
  <div class="not-found">
    <el-result icon="warning" title="Page not found" sub-title="The current path does not match a MockServer console page">
      <template #extra>
        <el-button type="primary" @click="$router.push('/rulesets')">Back to rulesets</el-button>
      </template>
    </el-result>
  </div>
</template>

<style lang="scss" scoped>
.not-found {
  height: 100%;
  display: grid;
  place-items: center;
}
</style>
