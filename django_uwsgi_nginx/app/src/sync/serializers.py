from rest_framework import serializers


class ErrorSerializer(serializers.Serializer):
    code = serializers.CharField(required=False, allow_null=True)
    message = serializers.CharField(required=False, allow_null=True)


class TrackDataSerializer(serializers.Serializer):
    priority = serializers.IntegerField(required=False, allow_null=True)
    is_system = serializers.BooleanField(required=False, default=False)


class StatusEventSerializer(serializers.Serializer):
    state = serializers.CharField(required=True)
    error = ErrorSerializer(required=False, allow_null=True)
    trackData = TrackDataSerializer(required=False, allow_null=True)
    updatedAt = serializers.CharField(required=True)
    txId = serializers.CharField(required=True)
    email = serializers.CharField(required=False, allow_null=True, allow_blank=True)
    channel_id = serializers.CharField(required=False, allow_null=True, allow_blank=True)
    channel = serializers.CharField(required=False, allow_null=True, allow_blank=True)
