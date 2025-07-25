package services

type AuthService struct {
    userRepo *repositories.UserRepository
    cfg      *config.Config
}

func NewAuthService(repo *repositories.UserRepository, cfg *config.Config) *AuthService {
    return &AuthService{userRepo: repo, cfg: cfg}
}
