from flask import Flask, request, jsonify
import joblib
import pandas as pd
import numpy as np

print("--- Servidor de IA do Sniper Bot ---")

# Carrega o cérebro que treinamos e salvamos anteriormente
try:
    model = joblib.load('anomaly_model.pkl')
    print("Modelo de IA 'anomaly_model.pkl' carregado com sucesso.")
except FileNotFoundError:
    print("ERRO: Arquivo 'anomaly_model.pkl' não encontrado!")
    print("Por favor, execute o notebook de treinamento primeiro para criar o modelo.")
    model = None

# Cria a aplicação da API
app = Flask(__name__)

# Define a rota '/predict' que aceitará perguntas
@app.route('/predict', methods=['POST'])
def predict():
    if model is None:
        return jsonify({'error': 'Modelo de IA não foi carregado.'}), 500

    try:
        data = request.get_json(force=True)
        
        # Prepara os dados recebidos do bot Go no formato que o modelo espera
        # Garante que os nomes e a ordem das colunas estão corretos
        features_df = pd.DataFrame([data], columns=['value', 'gas_limit', 'input_data_size'])
        
        # Faz a previsão e obtém a probabilidade
        prediction = model.predict(features_df)
        
        # Retorna o resultado em um formato JSON claro
        result = {
            'is_anomaly': int(prediction[0]) # -1 para anomalia, 1 para normal
        }
        return jsonify(result)
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    # Inicia o servidor na porta 5000, acessível na sua rede local
    app.run(host='0.0.0.0', port=5000, debug=False)