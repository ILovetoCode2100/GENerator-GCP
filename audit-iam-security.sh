#!/bin/bash

echo "🔍 IAM Security Audit for CDK Deployment"
echo "========================================"

USER_NAME="virtuoso-cdk-deploy"
POLICY_NAME="VirtuosoCDKDeploymentPolicy"

# Check if user exists
if ! aws iam get-user --user-name "$USER_NAME" >/dev/null 2>&1; then
    echo "❌ User '$USER_NAME' does not exist"
    exit 1
fi

echo "👤 Auditing user: $USER_NAME"
echo ""

# Get user details
USER_INFO=$(aws iam get-user --user-name "$USER_NAME")
CREATED_DATE=$(echo "$USER_INFO" | jq -r '.User.CreateDate')
echo "📅 User created: $CREATED_DATE"

# Check last used
LAST_USED=$(aws iam get-user --user-name "$USER_NAME" --query 'User.PasswordLastUsed' --output text 2>/dev/null)
if [[ "$LAST_USED" != "None" ]] && [[ "$LAST_USED" != "null" ]]; then
    echo "🕒 Password last used: $LAST_USED"
else
    echo "🕒 Password last used: Never (API keys only)"
fi

# Check access keys
echo ""
echo "🔑 Access Key Analysis:"
echo "======================"
ACCESS_KEYS=$(aws iam list-access-keys --user-name "$USER_NAME")
KEY_COUNT=$(echo "$ACCESS_KEYS" | jq '.AccessKeyMetadata | length')
echo "📊 Number of access keys: $KEY_COUNT"

if [[ $KEY_COUNT -gt 0 ]]; then
    echo "$ACCESS_KEYS" | jq -r '.AccessKeyMetadata[] | "Key ID: \(.AccessKeyId) | Status: \(.Status) | Created: \(.CreateDate)"'
    
    # Check key age
    echo ""
    echo "⏰ Key Age Analysis:"
    CURRENT_DATE=$(date +%s)
    echo "$ACCESS_KEYS" | jq -r '.AccessKeyMetadata[]' | while IFS= read -r key; do
        KEY_ID=$(echo "$key" | jq -r '.AccessKeyId')
        CREATED=$(echo "$key" | jq -r '.CreateDate')
        CREATED_TIMESTAMP=$(date -d "$CREATED" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "${CREATED%Z}" +%s 2>/dev/null)
        
        if [[ -n "$CREATED_TIMESTAMP" ]]; then
            AGE_DAYS=$(( (CURRENT_DATE - CREATED_TIMESTAMP) / 86400 ))
            if [[ $AGE_DAYS -gt 90 ]]; then
                echo "⚠️  Key $KEY_ID is $AGE_DAYS days old (>90 days - consider rotation)"
            elif [[ $AGE_DAYS -gt 60 ]]; then
                echo "🟡 Key $KEY_ID is $AGE_DAYS days old (>60 days - plan rotation)"
            else
                echo "✅ Key $KEY_ID is $AGE_DAYS days old (acceptable)"
            fi
        fi
    done
    
    # Check last used for each key
    echo ""
    echo "🕒 Key Usage Analysis:"
    echo "$ACCESS_KEYS" | jq -r '.AccessKeyMetadata[].AccessKeyId' | while read -r key_id; do
        LAST_USED_INFO=$(aws iam get-access-key-last-used --access-key-id "$key_id" 2>/dev/null)
        if [[ $? -eq 0 ]]; then
            LAST_USED_DATE=$(echo "$LAST_USED_INFO" | jq -r '.AccessKeyLastUsed.LastUsedDate // "Never"')
            SERVICE=$(echo "$LAST_USED_INFO" | jq -r '.AccessKeyLastUsed.ServiceName // "N/A"')
            REGION=$(echo "$LAST_USED_INFO" | jq -r '.AccessKeyLastUsed.Region // "N/A"')
            
            if [[ "$LAST_USED_DATE" != "Never" ]] && [[ "$LAST_USED_DATE" != "null" ]]; then
                echo "Key $key_id: Last used $LAST_USED_DATE ($SERVICE in $REGION)"
            else
                echo "Key $key_id: Never used"
            fi
        fi
    done
fi

# Check attached policies
echo ""
echo "📜 Policy Analysis:"
echo "=================="
ATTACHED_POLICIES=$(aws iam list-attached-user-policies --user-name "$USER_NAME")
POLICY_COUNT=$(echo "$ATTACHED_POLICIES" | jq '.AttachedPolicies | length')
echo "📊 Number of attached policies: $POLICY_COUNT"

if [[ $POLICY_COUNT -gt 0 ]]; then
    echo "$ATTACHED_POLICIES" | jq -r '.AttachedPolicies[] | "Policy: \(.PolicyName) | ARN: \(.PolicyArn)"'
    
    # Check if our custom policy is attached
    if echo "$ATTACHED_POLICIES" | jq -e ".AttachedPolicies[] | select(.PolicyName == \"$POLICY_NAME\")" >/dev/null; then
        echo "✅ Custom policy '$POLICY_NAME' is properly attached"
    else
        echo "⚠️  Custom policy '$POLICY_NAME' is NOT attached"
    fi
fi

# Check inline policies
INLINE_POLICIES=$(aws iam list-user-policies --user-name "$USER_NAME")
INLINE_COUNT=$(echo "$INLINE_POLICIES" | jq '.PolicyNames | length')
if [[ $INLINE_COUNT -gt 0 ]]; then
    echo "⚠️  User has $INLINE_COUNT inline policies (not recommended):"
    echo "$INLINE_POLICIES" | jq -r '.PolicyNames[]'
else
    echo "✅ No inline policies (good practice)"
fi

# Check groups
echo ""
echo "👥 Group Membership:"
echo "==================="
GROUPS=$(aws iam get-groups-for-user --user-name "$USER_NAME")
GROUP_COUNT=$(echo "$GROUPS" | jq '.Groups | length')
if [[ $GROUP_COUNT -gt 0 ]]; then
    echo "📊 Member of $GROUP_COUNT groups:"
    echo "$GROUPS" | jq -r '.Groups[] | "Group: \(.GroupName) | Created: \(.CreateDate)"'
else
    echo "✅ Not a member of any groups (expected for service accounts)"
fi

# Security recommendations
echo ""
echo "🔐 Security Assessment:"
echo "======================"

ISSUES=0
WARNINGS=0

# Check for MFA
MFA_DEVICES=$(aws iam list-mfa-devices --user-name "$USER_NAME")
MFA_COUNT=$(echo "$MFA_DEVICES" | jq '.MFADevices | length')
if [[ $MFA_COUNT -eq 0 ]]; then
    echo "ℹ️  No MFA devices (acceptable for service accounts)"
else
    echo "🔒 MFA devices configured: $MFA_COUNT"
fi

# Check for console access
if aws iam get-login-profile --user-name "$USER_NAME" >/dev/null 2>&1; then
    echo "⚠️  User has console access (not recommended for service accounts)"
    ((WARNINGS++))
else
    echo "✅ No console access (good for service accounts)"
fi

# Check key age
if [[ $KEY_COUNT -gt 0 ]]; then
    OLD_KEYS=$(aws iam list-access-keys --user-name "$USER_NAME" | jq -r '.AccessKeyMetadata[]' | while IFS= read -r key; do
        CREATED=$(echo "$key" | jq -r '.CreateDate')
        CREATED_TIMESTAMP=$(date -d "$CREATED" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "${CREATED%Z}" +%s 2>/dev/null)
        
        if [[ -n "$CREATED_TIMESTAMP" ]]; then
            AGE_DAYS=$(( ($(date +%s) - CREATED_TIMESTAMP) / 86400 ))
            if [[ $AGE_DAYS -gt 90 ]]; then
                echo "old"
            fi
        fi
    done)
    
    if [[ -n "$OLD_KEYS" ]]; then
        echo "⚠️  Some access keys are over 90 days old"
        ((WARNINGS++))
    fi
fi

# Check for unused keys
UNUSED_KEYS=$(aws iam list-access-keys --user-name "$USER_NAME" | jq -r '.AccessKeyMetadata[].AccessKeyId' | while read -r key_id; do
    LAST_USED_INFO=$(aws iam get-access-key-last-used --access-key-id "$key_id" 2>/dev/null)
    LAST_USED_DATE=$(echo "$LAST_USED_INFO" | jq -r '.AccessKeyLastUsed.LastUsedDate // "Never"')
    
    if [[ "$LAST_USED_DATE" == "Never" ]] || [[ "$LAST_USED_DATE" == "null" ]]; then
        echo "unused"
    fi
done)

if [[ -n "$UNUSED_KEYS" ]]; then
    echo "ℹ️  Some access keys have never been used (may be newly created)"
fi

echo ""
echo "📊 Security Score Summary:"
echo "========================="
if [[ $ISSUES -eq 0 ]] && [[ $WARNINGS -eq 0 ]]; then
    echo "🟢 EXCELLENT: No security issues found"
elif [[ $ISSUES -eq 0 ]] && [[ $WARNINGS -le 2 ]]; then
    echo "🟡 GOOD: Minor warnings found ($WARNINGS)"
elif [[ $ISSUES -eq 0 ]]; then
    echo "🟠 FAIR: Multiple warnings found ($WARNINGS)"
else
    echo "🔴 POOR: Security issues found (Issues: $ISSUES, Warnings: $WARNINGS)"
fi

echo ""
echo "📋 Recommendations:"
echo "=================="
echo "• Rotate access keys every 90 days"
echo "• Monitor CloudTrail logs for unusual activity"
echo "• Review permissions quarterly"
echo "• Use temporary credentials (STS) when possible"
echo "• Consider AWS SSO for human users"
echo "• Enable AWS Config for compliance monitoring"

echo ""
echo "✅ Security audit completed!"