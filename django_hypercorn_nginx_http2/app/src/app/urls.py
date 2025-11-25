from django.urls import include, path

urlpatterns = [
    path('', include('sync.urls')),
]
