from django.conf.urls import url
from django.urls import include
from rest_framework import routers

from sync.views import StatusViewSet

router = routers.DefaultRouter()
router.register('status', StatusViewSet, basename='status')

urlpatterns = [
    url(r'^', include(router.urls)),
]
